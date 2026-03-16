import {
  JsonRpcProvider,
  Wallet,
  Contract,
  ContractFactory,
  parseEther,
  formatEther,
  getAddress,
  InterfaceAbi,
  TransactionReceipt,
  TransactionResponse,
} from "ethers";

import {
  REGISTRY_ADDRESS,
  REPUTATION_ADDRESS,
  WALLET_ADDRESS,
  REGISTRY_ABI,
  REPUTATION_ABI,
  WALLET_ABI,
  TRUST_FULL,
  TrustLevel,
  trustLevelName,
} from "./precompiles";

const PRECOMPILE_TX_GAS_LIMIT = 2_000_000n;

// ─── Types ──────────────────────────────────────────────────────────

export interface AgentInfo {
  agentId: string;
  capabilities: string[];
  model: string;
  reputation: bigint;
  isOnline: boolean;
}

export interface TrustInfo {
  level: number;
  levelName: string;
  txLimit: string;
  dailyLimit: string;
  authorizedAt: bigint;
  expiresAt: bigint;
}

export interface WalletInfo {
  txLimit: string;
  dailyLimit: string;
  dailySpent: string;
  isFrozen: boolean;
  owner: string;
  operator: string;
  guardian: string;
}

// ─── Client ─────────────────────────────────────────────────────────

export class AgentClient {
  readonly provider: JsonRpcProvider;
  private _signer: Wallet | null = null;

  private _registry: Contract;
  private _reputation: Contract;
  private _wallet: Contract;

  constructor(rpcUrl: string, privateKey?: string) {
    this.provider = new JsonRpcProvider(rpcUrl);

    if (privateKey) {
      this._signer = new Wallet(privateKey, this.provider);
    }

    const runner = this._signer ?? this.provider;
    this._registry = new Contract(REGISTRY_ADDRESS, REGISTRY_ABI, runner);
    this._reputation = new Contract(REPUTATION_ADDRESS, REPUTATION_ABI, runner);
    this._wallet = new Contract(WALLET_ADDRESS, WALLET_ABI, runner);
  }

  // ─── Connection helpers ─────────────────────────────────────────

  connect(privateKey: string): void {
    this._signer = new Wallet(privateKey, this.provider);
    const signer = this._signer;
    this._registry = new Contract(REGISTRY_ADDRESS, REGISTRY_ABI, signer);
    this._reputation = new Contract(REPUTATION_ADDRESS, REPUTATION_ABI, signer);
    this._wallet = new Contract(WALLET_ADDRESS, WALLET_ABI, signer);
  }

  get address(): string | null {
    return this._signer?.address ?? null;
  }

  async getChainId(): Promise<bigint> {
    const network = await this.provider.getNetwork();
    return network.chainId;
  }

  async getBlockNumber(): Promise<number> {
    return this.provider.getBlockNumber();
  }

  // ─── Balance ────────────────────────────────────────────────────

  async balance(address?: string): Promise<string> {
    const addr = address ?? this.address;
    if (!addr) throw new Error("No address specified");
    const wei = await this.provider.getBalance(getAddress(addr));
    return formatEther(wei);
  }

  async balanceWei(address?: string): Promise<bigint> {
    const addr = address ?? this.address;
    if (!addr) throw new Error("No address specified");
    return this.provider.getBalance(getAddress(addr));
  }

  // ─── Agent Registry (0x...0801) ─────────────────────────────────

  async isAgent(address: string): Promise<boolean> {
    return this._registry.isAgent(getAddress(address));
  }

  async getAgent(address: string): Promise<AgentInfo> {
    const r = await this._registry.getAgent(getAddress(address));
    return {
      agentId: r[0],
      capabilities: r[1],
      model: r[2],
      reputation: r[3],
      isOnline: r[4],
    };
  }

  async registerAgent(
    capabilities: string,
    model: string,
    stakeAxon: string = "100"
  ): Promise<TransactionResponse> {
    this._requireSigner();
    const stakeWei = parseEther(stakeAxon);
    return this._registry.register(
      capabilities,
      model,
      this._precompileTxOverrides(stakeWei)
    );
  }

  async addStake(stakeAxon: string): Promise<TransactionResponse> {
    this._requireSigner();
    return this._registry.addStake(
      this._precompileTxOverrides(parseEther(stakeAxon))
    );
  }

  async updateAgent(
    capabilities: string,
    model: string
  ): Promise<TransactionResponse> {
    this._requireSigner();
    return this._registry.updateAgent(
      capabilities,
      model,
      this._precompileTxOverrides()
    );
  }

  async heartbeat(): Promise<TransactionResponse> {
    this._requireSigner();
    return this._registry.heartbeat(this._precompileTxOverrides());
  }

  async deregister(): Promise<TransactionResponse> {
    this._requireSigner();
    return this._registry.deregister(this._precompileTxOverrides());
  }

  // ─── Reputation (0x...0802) ─────────────────────────────────────

  async getReputation(address: string): Promise<bigint> {
    return this._reputation.getReputation(getAddress(address));
  }

  async getReputations(addresses: string[]): Promise<bigint[]> {
    const addrs = addresses.map((a) => getAddress(a));
    return this._reputation.getReputations(addrs);
  }

  async meetsReputation(address: string, minRep: number): Promise<boolean> {
    return this._reputation.meetsReputation(getAddress(address), minRep);
  }

  // ─── Wallet (0x...0803) ─────────────────────────────────────────

  async createWallet(
    operator: string,
    guardian: string,
    txLimitAxon: string = "10",
    dailyLimitAxon: string = "100",
    cooldownBlocks: number = 10
  ): Promise<TransactionResponse> {
    this._requireSigner();
    return this._wallet.createWallet(
      getAddress(operator),
      getAddress(guardian),
      parseEther(txLimitAxon),
      parseEther(dailyLimitAxon),
      cooldownBlocks,
      this._precompileTxOverrides()
    );
  }

  async executeWallet(
    wallet: string,
    target: string,
    valueAxon: string = "0",
    data: string = "0x"
  ): Promise<TransactionResponse> {
    this._requireSigner();
    return this._wallet.execute(
      getAddress(wallet),
      getAddress(target),
      parseEther(valueAxon),
      data,
      this._precompileTxOverrides()
    );
  }

  async freezeWallet(wallet: string): Promise<TransactionResponse> {
    this._requireSigner();
    return this._wallet.freeze(getAddress(wallet), this._precompileTxOverrides());
  }

  async recoverWallet(
    wallet: string,
    newOperator: string
  ): Promise<TransactionResponse> {
    this._requireSigner();
    return this._wallet.recover(
      getAddress(wallet),
      getAddress(newOperator),
      this._precompileTxOverrides()
    );
  }

  async setTrust(
    wallet: string,
    target: string,
    level: TrustLevel = TRUST_FULL,
    txLimitAxon: string = "0",
    dailyLimitAxon: string = "0",
    expiresAt: number = 0
  ): Promise<TransactionResponse> {
    this._requireSigner();
    return this._wallet.setTrust(
      getAddress(wallet),
      getAddress(target),
      level,
      parseEther(txLimitAxon),
      parseEther(dailyLimitAxon),
      expiresAt,
      this._precompileTxOverrides()
    );
  }

  async removeTrust(
    wallet: string,
    target: string
  ): Promise<TransactionResponse> {
    this._requireSigner();
    return this._wallet.removeTrust(
      getAddress(wallet),
      getAddress(target),
      this._precompileTxOverrides()
    );
  }

  async getTrust(wallet: string, target: string): Promise<TrustInfo> {
    const r = await this._wallet.getTrust(
      getAddress(wallet),
      getAddress(target)
    );
    return {
      level: Number(r[0]),
      levelName: trustLevelName(Number(r[0])),
      txLimit: formatEther(r[1]),
      dailyLimit: formatEther(r[2]),
      authorizedAt: r[3],
      expiresAt: r[4],
    };
  }

  async getWalletInfo(walletAddress: string): Promise<WalletInfo> {
    const r = await this._wallet.getWalletInfo(getAddress(walletAddress));
    return {
      txLimit: formatEther(r[0]),
      dailyLimit: formatEther(r[1]),
      dailySpent: formatEther(r[2]),
      isFrozen: r[3],
      owner: r[4],
      operator: r[5],
      guardian: r[6],
    };
  }

  // ─── Smart Contract deployment ──────────────────────────────────

  async deployContract(
    bytecode: string,
    abi: InterfaceAbi = [],
    gas?: number
  ): Promise<{ txHash: string; contractAddress: string }> {
    this._requireSigner();
    const factory = new ContractFactory(abi, bytecode, this._signer!);
    const contract = await factory.deploy({ gasLimit: gas ?? 3_000_000 });
    const tx = contract.deploymentTransaction()!;
    const receipt = await tx.wait();
    return {
      txHash: tx.hash,
      contractAddress: receipt!.contractAddress!,
    };
  }

  async callContract(
    address: string,
    abi: InterfaceAbi,
    method: string,
    args: unknown[] = []
  ): Promise<unknown> {
    const contract = new Contract(
      getAddress(address),
      abi,
      this._signer ?? this.provider
    );
    return contract[method](...args);
  }

  async sendContractTx(
    address: string,
    abi: InterfaceAbi,
    method: string,
    args: unknown[] = [],
    value: bigint = 0n
  ): Promise<TransactionResponse> {
    this._requireSigner();
    const contract = new Contract(getAddress(address), abi, this._signer!);
    return contract[method](...args, { value });
  }

  // ─── Transfer ───────────────────────────────────────────────────

  async transfer(to: string, amountAxon: string): Promise<TransactionResponse> {
    this._requireSigner();
    return this._signer!.sendTransaction({
      to: getAddress(to),
      value: parseEther(amountAxon),
    });
  }

  // ─── Utilities ──────────────────────────────────────────────────

  async waitForTx(txHash: string): Promise<TransactionReceipt | null> {
    return this.provider.waitForTransaction(txHash);
  }

  // ─── Internals ──────────────────────────────────────────────────

  private _requireSigner(): asserts this is { _signer: Wallet } {
    if (!this._signer) {
      throw new Error(
        "No signer set. Pass a privateKey to the constructor or call connect()."
      );
    }
  }

  private _precompileTxOverrides(
    value?: bigint
  ): { gasLimit: bigint; value?: bigint } {
    const overrides: { gasLimit: bigint; value?: bigint } = {
      gasLimit: PRECOMPILE_TX_GAS_LIMIT,
    };
    if (value !== undefined) {
      overrides.value = value;
    }
    return overrides;
  }
}
