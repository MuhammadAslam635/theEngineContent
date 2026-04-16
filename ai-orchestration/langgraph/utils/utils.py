import os
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.output_parsers.openai_functions import JsonOutputFunctionsParser
from ..state.state import ContentengineState

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

async def run_agent_node(state: ContentengineState, agent, name: str):
    """Utility to run an agent node and format the output for the graph state."""
    result = await agent.ainvoke(state)
    return {
        "messages": [result],
        "usage_stats": {name: 1} # Placeholder for actual usage tracking
    }
