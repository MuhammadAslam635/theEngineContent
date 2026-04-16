from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from .tools import get_lead_enrichment_tools

def get_lead_enrichment_agent(llm: ChatOpenAI):
    tools = get_lead_enrichment_tools()
    prompt = ChatPromptTemplate.from_messages([
        ("system", "You are the ContentEngine Lead Enrichment Agent. Use your tools to find and enrich company information for tenant {tenant_id}."),
        MessagesPlaceholder(variable_name="messages"),
    ])
    return prompt | llm.bind_tools(tools)
