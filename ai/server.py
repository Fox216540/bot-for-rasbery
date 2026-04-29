from typing import Dict, List

import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

app = FastAPI()
model = SentenceTransformer("all-MiniLM-L6-v2")

notes: Dict[str, str] = {}
embeddings: Dict[str, List[float]] = {}


class AddRequest(BaseModel):
    id: str
    text: str


class DeleteRequest(BaseModel):
    id: str


class SearchRequest(BaseModel):
    query: str


class SearchResult(BaseModel):
    id: str
    score: float
    text: str


def cosine(a: np.ndarray, b: np.ndarray) -> float:
    denom = np.linalg.norm(a) * np.linalg.norm(b)
    if denom == 0:
        return 0.0
    return float(np.dot(a, b) / denom)


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.post("/add")
def add_note(payload: AddRequest) -> dict:
    note_id = payload.id.strip()
    text = payload.text.strip()
    if not note_id:
        raise HTTPException(status_code=400, detail="id is required")
    if not text:
        raise HTTPException(status_code=400, detail="text is required")

    emb = model.encode(text)
    notes[note_id] = text
    embeddings[note_id] = emb.tolist()
    return {"status": "ok", "count": len(notes)}


@app.post("/delete")
def delete_note(payload: DeleteRequest) -> dict:
    note_id = payload.id.strip()
    if not note_id:
        raise HTTPException(status_code=400, detail="id is required")

    existed = note_id in notes
    notes.pop(note_id, None)
    embeddings.pop(note_id, None)
    return {"status": "ok", "deleted": existed}


@app.post("/search", response_model=List[SearchResult])
def search(payload: SearchRequest) -> List[SearchResult]:
    query = payload.query.strip()
    if not query:
        raise HTTPException(status_code=400, detail="query is required")

    if not notes:
        return []

    q_emb = model.encode(query)
    results: List[SearchResult] = []
    for note_id, emb in embeddings.items():
        score = cosine(q_emb, np.array(emb))
        results.append(SearchResult(id=note_id, score=score, text=notes.get(note_id, "")))

    results.sort(key=lambda x: x.score, reverse=True)
    return results[:5]
