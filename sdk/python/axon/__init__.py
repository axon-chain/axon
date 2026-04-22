from axon.client import AgentClient
from axon.ai_challenge import AIChallengeClient, get_reputation, get_agent_info
from axon.precompiles import (
    REGISTRY_ADDRESS, REPUTATION_ADDRESS, WALLET_ADDRESS,
    TRUST_BLOCKED, TRUST_UNKNOWN, TRUST_LIMITED, TRUST_FULL,
)

__version__ = "0.4.0"
__all__ = [
    "AgentClient",
    "AIChallengeClient",
    "get_reputation",
    "get_agent_info",
    "REGISTRY_ADDRESS", "REPUTATION_ADDRESS", "WALLET_ADDRESS",
    "TRUST_BLOCKED", "TRUST_UNKNOWN", "TRUST_LIMITED", "TRUST_FULL",
]
