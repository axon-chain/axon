"""Axon AI Challenge Client — Python SDK for AI Challenge participation."""

from typing import Optional, Dict, Any
from web3 import Web3
import hashlib
import json


class AIChallengeClient:
    """Client for participating in Axon AI Challenges.
    
    Usage::
    
        challenge = AIChallengeClient("https://mainnet-rpc.axonchain.ai/")
        challenge.set_account("0x...")
        
        # Get current challenge
        current = challenge.get_active_challenge()
        
        # Submit answer (hashed)
        challenge.submit_answer(challenge_id, answer)
        
        # Reveal during window
        challenge.reveal_answer(challenge_id, plain_answer)
    """
    
    # AI Challenge uses SHA256 for answer normalization
    # normalizeAnswer: lowercase ASCII letters, remove spaces/tabs/newlines
    
    def __init__(self, rpc_url: str):
        self.w3 = Web3(Web3.HTTPProvider(rpc_url))
        if not self.w3.is_connected():
            raise ConnectionError(f"Cannot connect to {rpc_url}")
        
        self._account = None
    
    @property
    def chain_id(self) -> int:
        return self.w3.eth.chain_id
    
    @property
    def block_number(self) -> int:
        return self.w3.eth.block_number
    
    def set_account(self, private_key: str):
        """Set account from private key."""
        from eth_account import Account
        self._account = Account.from_key(private_key)
    
    @property
    def address(self) -> Optional[str]:
        return self._account.address if self._account else None
    
    @staticmethod
    def normalize_answer(answer: str) -> str:
        """Normalize answer per AI Challenge rules.
        
        - Lowercase ASCII letters
        - Remove spaces, tabs, newlines
        """
        result = []
        for char in answer:
            if char in ' \t\n':
                continue
            if char.isascii() and char.isalpha():
                result.append(char.lower())
            else:
                result.append(char)
        return ''.join(result)
    
    @staticmethod
    def hash_answer(answer: str) -> bytes:
        """Hash normalized answer with SHA256."""
        normalized = AIChallengeClient.normalize_answer(answer)
        return self.w3.keccak(text=normalized)
    
    def get_active_challenge(self) -> Dict[str, Any]:
        """Get the current active challenge from latest block.
        
        Returns dict with:
        - challenge_id: int
        - question: str
        - answer_hash: bytes32
        - submit_deadline: int
        - reveal_deadline: int
        - block_number: int
        """
        # Query latest blocks for AIChallenge event
        # This is a simplified version - real implementation would index events
        current_block = self.w3.eth.block_number
        return {
            'challenge_id': 0,
            'question': 'Sample: What is 2+2?',
            'answer_hash': b'\x00' * 32,
            'submit_deadline': current_block + 100,
            'reveal_deadline': current_block + 150,
            'block_number': current_block,
        }
    
    def submit_answer(self, challenge_id: int, answer: str) -> str:
        """Submit hashed answer to a challenge.
        
        Args:
            challenge_id: ID of the challenge
            answer: Plain text answer
            
        Returns:
            Transaction hash
        """
        if not self._account:
            raise ValueError("Account not set. Call set_account() first.")
        
        answer_hash = self.hash_answer(answer)
        
        # Build transaction - this would call the AI Challenge contract
        # Actual contract address needs to be found in the chain
        # Placeholder: would be the AI Challenge module
        tx = {
            'from': self._account.address,
            'nonce': self.w3.eth.get_transaction_count(self._account.address),
            'gas': 200000,
            'gasPrice': self.w3.eth.gas_price,
            'chainId': self.chain_id,
        }
        
        # Sign and send (placeholder - actual contract call)
        signed = self._account.sign_transaction(tx)
        tx_hash = self.w3.eth.send_raw_transaction(signed.rawTransaction)
        
        return tx_hash.hex()
    
    def reveal_answer(self, challenge_id: int, plain_answer: str) -> str:
        """Reveal plain answer during reveal window.
        
        Args:
            challenge_id: ID of the challenge  
            plain_answer: The plain text answer to reveal
            
        Returns:
            Transaction hash
        """
        if not self._account:
            raise ValueError("Account not set. Call set_account() first.")
        
        tx = {
            'from': self._account.address,
            'nonce': self.w3.eth.get_transaction_count(self._account.address),
            'gas': 200000,
            'gasPrice': self.w3.eth.gas_price,
            'chainId': self.chain_id,
        }
        
        signed = self._account.sign_transaction(tx)
        tx_hash = self.w3.eth.send_raw_transaction(signed.rawTransaction)
        
        return tx_hash.hex()
    
    def get_challenge_result(self, challenge_id: int) -> Dict[str, Any]:
        """Get result of a past challenge.
        
        Returns:
            - correct_answer: str
            - participants: int
            - rewards_distributed: bool
        """
        # Would query challenge results from chain
        return {
            'correct_answer': '4',
            'participants': 0,
            'rewards_distributed': False,
        }


# Helper for standalone usage
def get_reputation(agent_address: str, rpc_url: str = "https://mainnet-rpc.axonchain.ai") -> int:
    """Quick helper to get agent reputation."""
    from axon.precompiles import REPUTATION_ADDRESS, REPUTATION_ABI
    
    w3 = Web3(Web3.HTTPProvider(rpc_url))
    contract = w3.eth.contract(
        address=Web3.to_checksum_address(REPUTATION_ADDRESS),
        abi=REPUTATION_ABI
    )
    
    result = contract.functions.getReputation(
        Web3.to_checksum_address(agent_address)
    ).call()
    
    return result


def get_agent_info(agent_address: str, rpc_url: str = "https://mainnet-rpc.axonchain.ai") -> Dict[str, Any]:
    """Quick helper to get agent info."""
    from axon.precompiles import REGISTRY_ADDRESS, REGISTRY_ABI
    
    w3 = Web3(Web3.HTTPProvider(rpc_url))
    contract = w3.eth.contract(
        address=Web3.to_checksum_address(REGISTRY_ADDRESS),
        abi=REGISTRY_ABI
    )
    
    result = contract.functions.getAgent(
        Web3.to_checksum_address(agent_address)
    ).call()
    
    return {
        'agent_id': result[0],
        'capabilities': result[1],
        'model': result[2],
        'reputation': result[3],
        'is_online': result[4],
    }