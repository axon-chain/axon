"""Axon Agent Client — Python SDK for the Axon blockchain."""

from typing import Optional, Tuple, List, Any, Dict
from web3 import Web3
from eth_account import Account

from axon.precompiles import (
    REGISTRY_ADDRESS, REPUTATION_ADDRESS, WALLET_ADDRESS,
    REGISTRY_ABI, REPUTATION_ABI, WALLET_ABI,
    TRUST_BLOCKED, TRUST_UNKNOWN, TRUST_LIMITED, TRUST_FULL,
)

AXON_DECIMALS = 18
ONE_AXON = 10 ** AXON_DECIMALS
PRECOMPILE_TX_GAS_LIMIT = 2_000_000


class AgentClient:
    """High-level client for interacting with the Axon chain.

    Usage::

        client = AgentClient("https://mainnet-rpc.axonchain.ai/")
        client.set_account("0x...")
        client.register_agent("nlp,reasoning", "axon-7b", stake_axon=100)
    """

    def __init__(self, rpc_url: str):
        self.w3 = Web3(Web3.HTTPProvider(rpc_url))
        if not self.w3.is_connected():
            raise ConnectionError(f"Cannot connect to {rpc_url}")

        self._account = None
        self._registry = self.w3.eth.contract(
            address=Web3.to_checksum_address(REGISTRY_ADDRESS), abi=REGISTRY_ABI
        )
        self._reputation = self.w3.eth.contract(
            address=Web3.to_checksum_address(REPUTATION_ADDRESS), abi=REPUTATION_ABI
        )
        self._wallet = self.w3.eth.contract(
            address=Web3.to_checksum_address(WALLET_ADDRESS), abi=WALLET_ABI
        )

    # ─── Connection info ─────────────────────────────────────────────

    @property
    def chain_id(self) -> int:
        return self.w3.eth.chain_id

    @property
    def block_number(self) -> int:
        return self.w3.eth.block_number

    @property
    def address(self) -> Optional[str]:
        return self._account.address if self._account else None

    def balance(self, address: Optional[str] = None) -> float:
        """Return AXON balance (human-readable, not wei)."""
        addr = address or self.address
        if not addr:
            raise ValueError("No address specified")
        wei = self.w3.eth.get_balance(Web3.to_checksum_address(addr))
        return wei / ONE_AXON

    def balance_wei(self, address: Optional[str] = None) -> int:
        """Return raw balance in aaxon (wei)."""
        addr = address or self.address
        if not addr:
            raise ValueError("No address specified")
        return self.w3.eth.get_balance(Web3.to_checksum_address(addr))

    # ─── Account management ──────────────────────────────────────────

    def set_account(self, private_key: str):
        """Set the signing account from a private key."""
        self._account = Account.from_key(private_key)

    def create_account(self) -> Tuple[str, str]:
        """Create a new random account. Returns (address, private_key)."""
        acct = Account.create()
        self._account = acct
        return acct.address, acct.key.hex()

    # ─── Agent Registry (0x...0801) ──────────────────────────────────

    def is_agent(self, address: str) -> bool:
        """Check if an address is a registered Agent."""
        return self._registry.functions.isAgent(
            Web3.to_checksum_address(address)
        ).call()

    def get_agent(self, address: str) -> Dict[str, Any]:
        """Query full Agent info from the registry."""
        result = self._registry.functions.getAgent(
            Web3.to_checksum_address(address)
        ).call()
        return {
            "agent_id": result[0],
            "capabilities": result[1],
            "model": result[2],
            "reputation": result[3],
            "is_online": result[4],
        }

    def register_agent(
        self, capabilities: str, model: str, stake_axon: int = 100
    ) -> str:
        """Register as an AI Agent. Requires >= 100 AXON stake (20 burned)."""
        self._require_account()
        stake_wei = int(stake_axon) * ONE_AXON
        tx = self._registry.functions.register(
            capabilities, model
        ).build_transaction(self._tx_params(value=stake_wei))
        return self._send_tx(tx)

    def add_stake(self, stake_axon: int) -> str:
        """Add more stake to an existing AI Agent."""
        self._require_account()
        stake_wei = int(stake_axon) * ONE_AXON
        tx = self._registry.functions.addStake().build_transaction(
            self._tx_params(value=stake_wei)
        )
        return self._send_tx(tx)

    def update_agent(self, capabilities: str, model: str) -> str:
        """Update Agent capabilities and model."""
        self._require_account()
        tx = self._registry.functions.updateAgent(
            capabilities, model
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def heartbeat(self) -> str:
        """Send a heartbeat to keep Agent status ONLINE."""
        self._require_account()
        tx = self._registry.functions.heartbeat().build_transaction(self._tx_params())
        return self._send_tx(tx)

    def deregister(self) -> str:
        """Start the deregistration process (stake returned after cooldown)."""
        self._require_account()
        tx = self._registry.functions.deregister().build_transaction(self._tx_params())
        return self._send_tx(tx)

    # ─── Reputation (0x...0802) ──────────────────────────────────────

    def get_reputation(self, address: str) -> int:
        """Get an Agent's reputation score (0-100)."""
        return self._reputation.functions.getReputation(
            Web3.to_checksum_address(address)
        ).call()

    def get_reputations(self, addresses: List[str]) -> List[int]:
        """Batch-query reputation scores for multiple Agents."""
        addrs = [Web3.to_checksum_address(a) for a in addresses]
        return self._reputation.functions.getReputations(addrs).call()

    def meets_reputation(self, address: str, min_rep: int) -> bool:
        """Check if an Agent meets a minimum reputation threshold."""
        return self._reputation.functions.meetsReputation(
            Web3.to_checksum_address(address), min_rep
        ).call()

    # ─── Wallet (0x...0803) ──────────────────────────────────────────

    def create_wallet(
        self,
        operator: str,
        guardian: str,
        tx_limit_axon: int = 10,
        daily_limit_axon: int = 100,
        cooldown_blocks: int = 10,
    ) -> str:
        """Create an Agent smart wallet. Caller becomes the Owner."""
        self._require_account()
        tx = self._wallet.functions.createWallet(
            Web3.to_checksum_address(operator),
            Web3.to_checksum_address(guardian),
            int(tx_limit_axon) * ONE_AXON,
            int(daily_limit_axon) * ONE_AXON,
            cooldown_blocks,
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def execute_wallet(
        self, wallet: str, target: str, value_axon: int = 0, data: bytes = b""
    ) -> str:
        """Execute a transaction through the Agent wallet (as Operator)."""
        self._require_account()
        tx = self._wallet.functions.execute(
            Web3.to_checksum_address(wallet),
            Web3.to_checksum_address(target),
            int(value_axon) * ONE_AXON,
            data,
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def freeze_wallet(self, wallet: str) -> str:
        """Freeze a wallet (Guardian or Owner only)."""
        self._require_account()
        tx = self._wallet.functions.freeze(
            Web3.to_checksum_address(wallet)
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def recover_wallet(self, wallet: str, new_operator: str) -> str:
        """Recover a wallet by replacing the Operator (Guardian only)."""
        self._require_account()
        tx = self._wallet.functions.recover(
            Web3.to_checksum_address(wallet),
            Web3.to_checksum_address(new_operator),
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def set_trust(
        self,
        wallet: str,
        target: str,
        level: int = TRUST_FULL,
        tx_limit_axon: int = 0,
        daily_limit_axon: int = 0,
        expires_at: int = 0,
    ) -> str:
        """Authorize a contract at a trust level (Owner only).

        Trust levels:
            0 = Blocked — always rejected
            1 = Unknown  — wallet default limits
            2 = Limited  — custom per-channel limits
            3 = Full     — no limits
        """
        self._require_account()
        tx = self._wallet.functions.setTrust(
            Web3.to_checksum_address(wallet),
            Web3.to_checksum_address(target),
            level,
            int(tx_limit_axon) * ONE_AXON,
            int(daily_limit_axon) * ONE_AXON,
            expires_at,
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def remove_trust(self, wallet: str, target: str) -> str:
        """Remove trust authorization for a contract (Owner only)."""
        self._require_account()
        tx = self._wallet.functions.removeTrust(
            Web3.to_checksum_address(wallet),
            Web3.to_checksum_address(target),
        ).build_transaction(self._tx_params())
        return self._send_tx(tx)

    def get_trust(self, wallet: str, target: str) -> Dict[str, Any]:
        """Query trust level and limits for a contract."""
        result = self._wallet.functions.getTrust(
            Web3.to_checksum_address(wallet),
            Web3.to_checksum_address(target),
        ).call()
        level_names = {0: "blocked", 1: "unknown", 2: "limited", 3: "full"}
        return {
            "level": result[0],
            "level_name": level_names.get(result[0], "unknown"),
            "tx_limit": result[1] / ONE_AXON,
            "daily_limit": result[2] / ONE_AXON,
            "authorized_at": result[3],
            "expires_at": result[4],
        }

    def get_wallet_info(self, wallet_address: str) -> Dict[str, Any]:
        """Query wallet configuration and status."""
        result = self._wallet.functions.getWalletInfo(
            Web3.to_checksum_address(wallet_address)
        ).call()
        return {
            "tx_limit": result[0] / ONE_AXON,
            "daily_limit": result[1] / ONE_AXON,
            "daily_spent": result[2] / ONE_AXON,
            "is_frozen": result[3],
            "owner": result[4],
            "operator": result[5],
            "guardian": result[6],
        }

    # ─── Smart Contract deployment ───────────────────────────────────

    def deploy_contract(self, bytecode: str, **kwargs) -> Tuple[str, str]:
        """Deploy an EVM contract. Returns (tx_hash, contract_address).

        Note: deploying burns 10 AXON in addition to gas fees.
        """
        self._require_account()
        tx = {
            **self._tx_params(),
            "data": bytecode if bytecode.startswith("0x") else f"0x{bytecode}",
            "gas": kwargs.get("gas", 3_000_000),
        }
        tx_hash = self._send_tx(tx)
        receipt = self.w3.eth.wait_for_transaction_receipt(tx_hash, timeout=60)
        return tx_hash, receipt.contractAddress

    def call_contract(
        self, address: str, abi: list, method: str, args: list = None, value: int = 0
    ) -> Any:
        """Call a read-only contract method."""
        contract = self.w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=abi
        )
        fn = getattr(contract.functions, method)
        return fn(*(args or [])).call()

    def send_contract_tx(
        self, address: str, abi: list, method: str, args: list = None, value: int = 0
    ) -> str:
        """Send a state-changing contract transaction."""
        self._require_account()
        contract = self.w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=abi
        )
        fn = getattr(contract.functions, method)
        tx = fn(*(args or [])).build_transaction(self._tx_params(value=value))
        return self._send_tx(tx)

    # ─── Transfer ────────────────────────────────────────────────────

    def transfer(self, to: str, amount_axon: int) -> str:
        """Send AXON to an address."""
        self._require_account()
        tx = {
            **self._tx_params(),
            "to": Web3.to_checksum_address(to),
            "value": int(amount_axon) * ONE_AXON,
            "gas": 21000,
        }
        return self._send_tx(tx)

    # ─── Internals ───────────────────────────────────────────────────

    def _require_account(self):
        if not self._account:
            raise ValueError("No account set. Call set_account() or create_account() first.")

    def _tx_params(self, value: int = 0) -> dict:
        return {
            "from": self._account.address,
            "nonce": self.w3.eth.get_transaction_count(self._account.address),
            "gas": PRECOMPILE_TX_GAS_LIMIT,
            "gasPrice": self.w3.eth.gas_price or 0,
            "chainId": self.chain_id,
            "value": value,
        }

    def _send_tx(self, tx: dict) -> str:
        signed = self._account.sign_transaction(tx)
        tx_hash = self.w3.eth.send_raw_transaction(signed.raw_transaction)
        return tx_hash.hex()

    def wait_for_tx(self, tx_hash: str, timeout: int = 30) -> dict:
        """Wait for a transaction to be mined and return the receipt."""
        receipt = self.w3.eth.wait_for_transaction_receipt(tx_hash, timeout=timeout)
        return dict(receipt)
