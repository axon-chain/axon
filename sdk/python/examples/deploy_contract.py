#!/usr/bin/env python3
"""
Example: Deploy an ERC-20 token contract on Axon.
"""

from axon import AgentClient

# Minimal ERC-20 bytecode (stores 42 in slot 0)
SIMPLE_STORAGE_BYTECODE = (
    "0x6080604052602a600055"
    "60158060156000396000f3fe"
    "60003560e01c80632a1afcd914600f57005b"
    "60005460405190815260200160405180910390f3"
)


def main():
    client = AgentClient("https://mainnet-rpc.axonchain.ai/")
    print(f"Chain ID: {client.chain_id}")

    # Set account (replace with your private key)
    # client.set_account("your_private_key_here")
    address, key = client.create_account()
    print(f"Deployer: {address}")

    bal = client.balance()
    print(f"Balance: {bal} AXON")

    if bal < 11:
        print("Need >= 11 AXON (10 burned on deploy + gas)")
        return

    print("\nDeploying contract...")
    tx_hash, contract_addr = client.deploy_contract(SIMPLE_STORAGE_BYTECODE)
    print(f"  TX: {tx_hash}")
    print(f"  Contract: {contract_addr}")

    print("\nDone!")


if __name__ == "__main__":
    main()
