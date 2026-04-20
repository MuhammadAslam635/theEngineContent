import os
import os
import sys
from typing import Any, Dict, Optional

from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.output_parsers.openai_functions import JsonOutputFunctionsParser
from langchain_core.messages import BaseMessage

from ..state.state import ContentengineState

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
from global_services.audit_logs_helpers import try_write_audit_log, normalize_audit_payload

def get_openrouter_llm(model_name: str = "anthropic/claude-3-sonnet"):
    """Initialize a ChatOpenAI instance configured for OpenRouter."""
    return ChatOpenAI(
        model=model_name,
        openai_api_key=os.getenv("OPENROUTER_API_KEY"),
        openai_api_base="https://openrouter.ai/api/v1",
        default_headers={
            "HTTP-Referer": "https://contentengine.ai", # Optional: replace with your site
            "X-Title": "ContentEngine Orchestration"
        }
    )

def create_supervisor_chain(llm: ChatOpenAI, system_prompt: str, members: list[str]):
    """Create a supervisor chain that routes to different agents."""
    options = ["FINISH"] + members
    function_def = {
        "name": "route",
        "description": "Select the next role.",
        "parameters": {
            "title": "routeSchema",
            "type": "object",
            "properties": {
                "next": {
                    "title": "Next",
                    "anyOf": [
                        {"enum": options},
                    ],
                }
            },
            "required": ["next"],
        },
    }
    
    prompt = ChatPromptTemplate.from_messages([
        ("system", system_prompt),
        MessagesPlaceholder(variable_name="messages"),
        ("system", "Given the conversation above, who should act next? Or should we FINISH? Select one of: {options}"),
    ]).partial(options=str(options), members=", ".join(members))
    
    return prompt | llm.bind_functions(functions=[function_def], function_call="route") | JsonOutputFunctionsParser()

def _extract_agent_io(
    state: ContentengineState, result: BaseMessage, name: str
) -> Dict[str, Any]:
    """Pull out whatever we can from the agent result for the audit row."""
    payload: Dict[str, Any] = {
        "agent": name,
        "user_id": state.get("user_id"),
        "input_query": None,
        "agent_prompt": None,
        "input_tokens": 0,
        "output_response": None,
        "output_usage_tokens": 0,
        "error": None,
        "line_number": None,
    }

    if state.get("messages"):
        last_user_msg = next(
            (m for m in reversed(state["messages"]) if getattr(m, "type", None) == "human"), None
        )
        if last_user_msg:
            payload["input_query"] = last_user_msg.content

    if hasattr(result, "content"):
        payload["output_response"] = result.content

    if hasattr(result, "usage_metadata") and result.usage_metadata:
        payload["input_tokens"] = result.usage_metadata.get("input_tokens", 0)
        payload["output_usage_tokens"] = result.usage_metadata.get("output_tokens", 0)

    return payload


async def run_agent_node(state: ContentengineState, agent, name: str):
    """Utility to run an agent node and format the output for the graph state."""
    try:
        result = await agent.ainvoke(state)
    except Exception as exc:
        # still attempt to log the failure, then re-raise so the graph sees it
        payload = {
            "agent": name,
            "user_id": state.get("user_id"),
            "error": str(exc),
        }
        try_write_audit_log(**normalize_audit_payload(payload))
        raise

    # normal path: log success and return state update
    payload = _extract_agent_io(state, result, name)
    try_write_audit_log(**normalize_audit_payload(payload))

    return {
        "messages": [result],
        "usage_stats": {name: 1},  # Placeholder for actual usage tracking
    }
