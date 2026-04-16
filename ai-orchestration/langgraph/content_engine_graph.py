from langgraph.graph import StateGraph, END
from .state.state import ContentengineState
from .supervisors.lead_enrichment.supervisor import get_lead_enrichment_supervisor
from .agents.lead_enrichment.agent import get_lead_enrichment_agent
from .utils.utils import run_agent_node, get_openrouter_llm
import functools

def create_content_engine_graph(model_name: str = "anthropic/claude-3-sonnet"):
    llm = get_openrouter_llm(model_name)
    
    workflow = StateGraph(ContentengineState)
    
    # Define Supervisors and Agents (Simplified example with one branch)
    le_supervisor = get_lead_enrichment_supervisor(llm)
    le_agent = get_lead_enrichment_agent(llm)
    
    # Add Nodes
    workflow.add_node("LeadEnrichmentSupervisor", le_supervisor)
    workflow.add_node("LeadEnrichmentAgent", functools.partial(run_agent_node, agent=le_agent, name="LeadEnrichmentAgent"))
    
    # Add Edges
    workflow.set_entry_point("LeadEnrichmentSupervisor")
    
    # Routing logic from supervisor
    workflow.add_conditional_edges(
        "LeadEnrichmentSupervisor",
        lambda x: x["next"],
        {
            "LeadEnrichmentAgent": "LeadEnrichmentAgent",
            "FINISH": END
        }
    )
    
    # Agents always go back to their supervisor
    workflow.add_edge("LeadEnrichmentAgent", "LeadEnrichmentSupervisor")
    
    return workflow.compile()
