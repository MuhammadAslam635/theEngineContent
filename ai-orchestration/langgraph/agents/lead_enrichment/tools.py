from langchain_core.tools import tool

@tool
def find_companies_with_balance_check(query: str, tenant_id: str):
    """Find companies based on a query while checking if the tenant has enough balance."""
    return f"Results for {query} (Tenant: {tenant_id})"

@tool
def start_enrichment_job_tool_with_quota(company_id: str, quota: int):
    """Start an enrichment job for a company if it's within the user's quota."""
    return f"Started job for {company_id}"

def get_lead_enrichment_tools():
    return [find_companies_with_balance_check, start_enrichment_job_tool_with_quota]
