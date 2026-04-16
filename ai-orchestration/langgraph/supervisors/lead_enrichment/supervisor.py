from langchain_openai import ChatOpenAI
from ...utils.utils import create_supervisor_chain

def get_lead_enrichment_supervisor(llm: ChatOpenAI):
    system_prompt = (
        "You are the ContentEngine Lead Enrichment Supervisor. Your task is to coordinate the enrichment of leads. "
        "You manage the LeadEnrichmentAgent."
    )
    members = ["LeadEnrichmentAgent"]
    return create_supervisor_chain(llm, system_prompt, members)
