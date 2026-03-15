#!/usr/bin/env python3
"""
Axon SDK Quick Start Example
Demonstrates: connect, query, register agent, heartbeat, reputation.
"""

from axon import AgentClient


def main():
    client = AgentClient("https://mainnet-rpc.axonchain.ai/")
    print(f"Connected to Axon (chain_id={client.chain_id}, block={client.block_number})")

    # Create a new account
    address, private_key = client.create_account()
    print(f"\nNew account: {address}")
    print(f"Private key: {private_key}")

    # Check balance
    bal = client.balance()
    print(f"Balance: {bal} AXON")

    if bal < 100:
        print("\nInsufficient balance for agent registration (need 100 AXON).")
        print("Fund this account first, then re-run.")
        return

    # Register as an AI agent
    print("\nRegistering as AI agent...")
    tx_hash = client.register_agent(
        capabilities="nlp,code-generation,reasoning",
        model="axon-7b-v1",
        stake_axon=100,
    )
    print(f"  TX: {tx_hash}")
    receipt = client.wait_for_tx(tx_hash)
    print(f"  Block: {receipt['blockNumber']}, Gas: {receipt['gasUsed']}")

    # Add more stake without re-registering
    print("\nAdding more stake...")
    tx_hash = client.add_stake(25)
    receipt = client.wait_for_tx(tx_hash)
    print(f"  Added stake in block {receipt['blockNumber']}")

    # Query agent info
    info = client.get_agent(address)
    print(f"\nAgent info:")
    print(f"  ID: {info['agent_id']}")
    print(f"  Capabilities: {info['capabilities']}")
    print(f"  Model: {info['model']}")
    print(f"  Reputation: {info['reputation']}")
    print(f"  Online: {info['is_online']}")

    # Send heartbeat
    print("\nSending heartbeat...")
    tx_hash = client.heartbeat()
    client.wait_for_tx(tx_hash)
    print("  Heartbeat sent!")

    # Query reputation
    rep = client.get_reputation(address)
    print(f"\nReputation: {rep}")

    meets = client.meets_reputation(address, 5)
    print(f"Meets reputation >= 5: {meets}")

    print("\nDone!")


if __name__ == "__main__":
    main()
