import asyncio
import logging
import os
from contextlib import asynccontextmanager
from typing import Awaitable, Callable

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from lotus_lamp import DeviceConfig, LotusLamp


logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("strip-service")


class ColorRequest(BaseModel):
    r: int = Field(ge=0, le=255)
    g: int = Field(ge=0, le=255)
    b: int = Field(ge=0, le=255)


class ValueRequest(BaseModel):
    value: int = Field(ge=0)


class ModeRequest(BaseModel):
    mode: int = Field(ge=0, le=212)


class TimerRequest(BaseModel):
    seconds: int = Field(gt=0)


class StripController:
    def __init__(self, address: str, name: str, reconnect_delay: float = 5.0) -> None:
        self._address = address
        self._name = name
        self._reconnect_delay = reconnect_delay
        self._lock = asyncio.Lock()
        self._lamp: LotusLamp | None = None
        self._connected = False
        self._powered = False
        self._timer_task: asyncio.Task[None] | None = None
        self._reconnect_task: asyncio.Task[None] | None = None
        self._stopped = asyncio.Event()

    async def start(self) -> None:
        self._stopped.clear()
        self._reconnect_task = asyncio.create_task(self._reconnect_loop())

    async def stop(self) -> None:
        self._stopped.set()
        if self._reconnect_task is not None:
            self._reconnect_task.cancel()
            await asyncio.gather(self._reconnect_task, return_exceptions=True)
            self._reconnect_task = None
        await self.cancel_timer()
        async with self._lock:
            await self._disconnect_locked()

    async def power_on(self) -> None:
        await self._run("power_on", lambda lamp: lamp.power_on())
        self._powered = True

    async def power_off(self) -> None:
        await self._run("power_off", lambda lamp: lamp.power_off())
        self._powered = False

    async def set_rgb(self, r: int, g: int, b: int) -> None:
        await self._run("set_rgb", lambda lamp: lamp.set_rgb(r, g, b))
        self._powered = True

    async def set_brightness(self, value: int) -> None:
        if value > 100:
            raise HTTPException(status_code=422, detail="brightness must be 0..100")
        await self._run("set_brightness", lambda lamp: lamp.set_brightness(value))

    async def set_speed(self, value: int) -> None:
        await self._run("set_speed", lambda lamp: lamp.set_speed(value))

    async def set_mode(self, mode: int) -> None:
        await self._run("set_mode", lambda lamp: lamp.set_animation(mode))
        self._powered = True

    async def set_timer(self, seconds: int) -> None:
        await self.cancel_timer()
        self._timer_task = asyncio.create_task(self._timer(seconds))
        logger.info("timer set seconds=%s", seconds)

    async def cancel_timer(self) -> None:
        if self._timer_task is None:
            return
        self._timer_task.cancel()
        await asyncio.gather(self._timer_task, return_exceptions=True)
        self._timer_task = None
        logger.info("timer canceled")

    def status(self) -> dict[str, bool]:
        return {"connected": self._connected, "powered": self._powered}

    async def _timer(self, seconds: int) -> None:
        try:
            await asyncio.sleep(seconds)
            await self.power_off()
        finally:
            self._timer_task = None

    async def _reconnect_loop(self) -> None:
        while not self._stopped.is_set():
            if not self._connected:
                async with self._lock:
                    if not self._connected:
                        try:
                            await self._connect_locked()
                        except Exception as err:
                            logger.warning("connect failed: %s", err)
            await asyncio.sleep(self._reconnect_delay)

    async def _run(self, name: str, fn: Callable[[LotusLamp], Awaitable[None]]) -> None:
        async with self._lock:
            try:
                lamp = await self._ensure_connected_locked()
                logger.info("command start name=%s", name)
                await fn(lamp)
                logger.info("command done name=%s", name)
                return
            except HTTPException:
                raise
            except Exception as err:
                logger.warning("command failed name=%s error=%s", name, err)
                await self._disconnect_locked()

            try:
                lamp = await self._ensure_connected_locked()
                logger.info("command retry name=%s", name)
                await fn(lamp)
                logger.info("command retry done name=%s", name)
            except Exception as err:
                await self._disconnect_locked()
                raise HTTPException(status_code=503, detail=f"strip command failed: {err}") from err

    async def _ensure_connected_locked(self) -> LotusLamp:
        if self._lamp is not None and self._connected:
            return self._lamp
        try:
            await self._connect_locked()
        except Exception as err:
            raise HTTPException(status_code=503, detail=f"strip unavailable: {err}") from err
        if self._lamp is None:
            raise HTTPException(status_code=503, detail="strip unavailable")
        return self._lamp

    async def _connect_locked(self) -> None:
        config = DeviceConfig(name=self._name, address=self._address)
        lamp = LotusLamp(device_config=config)
        logger.info("connect start address=%s name=%s", self._address, self._name)
        await lamp.connect()
        self._lamp = lamp
        self._connected = True
        logger.info("connect done address=%s name=%s", self._address, self._name)

    async def _disconnect_locked(self) -> None:
        lamp = self._lamp
        self._lamp = None
        self._connected = False
        if lamp is None:
            return
        try:
            await lamp.disconnect()
            logger.info("disconnect done")
        except Exception as err:
            logger.warning("disconnect failed: %s", err)


strip = StripController(
    address=os.getenv("STRIP_ADDRESS", "BE:68:E6:00:53:2D"),
    name=os.getenv("STRIP_NAME", "ELK-BLEDDM 8C"),
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    await strip.start()
    try:
        yield
    finally:
        await strip.stop()


app = FastAPI(lifespan=lifespan)


@app.post("/power/on")
async def power_on() -> dict[str, bool]:
    await strip.power_on()
    return {"ok": True}


@app.post("/power/off")
async def power_off() -> dict[str, bool]:
    await strip.power_off()
    return {"ok": True}


@app.post("/color")
async def color(req: ColorRequest) -> dict[str, bool]:
    await strip.set_rgb(req.r, req.g, req.b)
    return {"ok": True}


@app.post("/brightness")
async def brightness(req: ValueRequest) -> dict[str, bool]:
    await strip.set_brightness(req.value)
    return {"ok": True}


@app.post("/speed")
async def speed(req: ValueRequest) -> dict[str, bool]:
    await strip.set_speed(req.value)
    return {"ok": True}


@app.post("/mode")
async def mode(req: ModeRequest) -> dict[str, bool]:
    await strip.set_mode(req.mode)
    return {"ok": True}


@app.post("/timer")
async def timer(req: TimerRequest) -> dict[str, bool]:
    await strip.set_timer(req.seconds)
    return {"ok": True}


@app.delete("/timer")
async def cancel_timer() -> dict[str, bool]:
    await strip.cancel_timer()
    return {"ok": True}


@app.get("/status")
async def status() -> dict[str, bool]:
    return strip.status()
