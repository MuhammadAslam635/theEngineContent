from typing import Sequence, Optional, List, Annotated, TypedDict
import operator
from langchain_core.messages import BaseMessage

def combine_usage(left: dict, right: dict) -> dict:
    """Merge usage statistics."""
    new_usage = left.copy()
    for k, v in right.items():
        new_usage[k] = new_usage.get(k, 0) + v
    return new_usage

def limit_list(last: int = 1000):
    """Return a function that limits a list to the last N items."""
    def merge(left: List[str], right: List[str]) -> List[str]:
        combined = left + right
        return combined[-last:]
    return merge

class ContentengineState(TypedDict):
    """The state of the orchestration graph."""
    messages: Annotated[Sequence[BaseMessage], operator.add]
    next: str
    job_id: Optional[str]
    user_id: str
    user_name: str
    tenant_id: str
    authorization: str
    usage_stats: Annotated[dict, combine_usage]
    valid_companies: Annotated[List[str], limit_list(1000)]
