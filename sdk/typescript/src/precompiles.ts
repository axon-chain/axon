export const REGISTRY_ADDRESS = "0x0000000000000000000000000000000000000801";
export const REPUTATION_ADDRESS = "0x0000000000000000000000000000000000000802";
export const WALLET_ADDRESS = "0x0000000000000000000000000000000000000803";

export const REGISTRY_ABI = [
  {
    inputs: [{ name: "account", type: "address" }],
    name: "isAgent",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "account", type: "address" }],
    name: "getAgent",
    outputs: [
      { name: "agentId", type: "string" },
      { name: "capabilities", type: "string[]" },
      { name: "model", type: "string" },
      { name: "reputation", type: "uint64" },
      { name: "isOnline", type: "bool" },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      { name: "capabilities", type: "string" },
      { name: "model", type: "string" },
    ],
    name: "register",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [],
    name: "addStake",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [
      { name: "capabilities", type: "string" },
      { name: "model", type: "string" },
    ],
    name: "updateAgent",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "heartbeat",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "deregister",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [{ name: "amount", type: "uint256" }],
    name: "reduceStake",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "claimReducedStake",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [{ name: "agent", type: "address" }],
    name: "getStakeInfo",
    outputs: [
      { name: "totalStake", type: "uint256" },
      { name: "pendingReduce", type: "uint256" },
      { name: "reduceUnlockHeight", type: "uint64" },
    ],
    stateMutability: "view",
    type: "function",
  },
] as const;

export const REPUTATION_ABI = [
  {
    inputs: [{ name: "agent", type: "address" }],
    name: "getReputation",
    outputs: [{ name: "", type: "uint64" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "agents", type: "address[]" }],
    name: "getReputations",
    outputs: [{ name: "", type: "uint64[]" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      { name: "agent", type: "address" },
      { name: "minReputation", type: "uint64" },
    ],
    name: "meetsReputation",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
] as const;

export const WALLET_ABI = [
  {
    inputs: [
      { name: "operator", type: "address" },
      { name: "guardian", type: "address" },
      { name: "txLimit", type: "uint256" },
      { name: "dailyLimit", type: "uint256" },
      { name: "cooldownBlocks", type: "uint256" },
    ],
    name: "createWallet",
    outputs: [{ name: "wallet", type: "address" }],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "wallet", type: "address" },
      { name: "target", type: "address" },
      { name: "value", type: "uint256" },
      { name: "data", type: "bytes" },
    ],
    name: "execute",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [{ name: "wallet", type: "address" }],
    name: "freeze",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "wallet", type: "address" },
      { name: "newOperator", type: "address" },
    ],
    name: "recover",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "wallet", type: "address" },
      { name: "target", type: "address" },
      { name: "level", type: "uint8" },
      { name: "txLimit", type: "uint256" },
      { name: "dailyLimit", type: "uint256" },
      { name: "expiresAt", type: "uint256" },
    ],
    name: "setTrust",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "wallet", type: "address" },
      { name: "target", type: "address" },
    ],
    name: "removeTrust",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "wallet", type: "address" },
      { name: "target", type: "address" },
    ],
    name: "getTrust",
    outputs: [
      { name: "level", type: "uint8" },
      { name: "txLimit", type: "uint256" },
      { name: "dailyLimit", type: "uint256" },
      { name: "authorizedAt", type: "uint256" },
      { name: "expiresAt", type: "uint256" },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "wallet", type: "address" }],
    name: "getWalletInfo",
    outputs: [
      { name: "txLimit", type: "uint256" },
      { name: "dailyLimit", type: "uint256" },
      { name: "dailySpent", type: "uint256" },
      { name: "isFrozen", type: "bool" },
      { name: "owner", type: "address" },
      { name: "operator", type: "address" },
      { name: "guardian", type: "address" },
    ],
    stateMutability: "view",
    type: "function",
  },
] as const;

// --- New precompile addresses (v2 upgrade) ---
export const REPORT_ADDRESS = "0x0000000000000000000000000000000000000807";
export const POSEIDON_ADDRESS = "0x0000000000000000000000000000000000000810";
export const PRIVATE_TRANSFER_ADDRESS = "0x0000000000000000000000000000000000000811";
export const PRIVATE_IDENTITY_ADDRESS = "0x0000000000000000000000000000000000000812";
export const ZK_VERIFIER_ADDRESS = "0x0000000000000000000000000000000000000813";

export const REPORT_ABI = [
  {
    inputs: [
      { name: "targetAgent", type: "address" },
      { name: "score", type: "int8" },
      { name: "evidence", type: "bytes32" },
      { name: "reason", type: "string" },
    ],
    name: "submitReport",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [{ name: "agent", type: "address" }],
    name: "getContractReputation",
    outputs: [
      { name: "score", type: "int64" },
      { name: "positiveCount", type: "uint64" },
      { name: "negativeCount", type: "uint64" },
      { name: "uniqueReporters", type: "uint64" },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "agent", type: "address" }],
    name: "getEpochReportCount",
    outputs: [{ name: "", type: "uint64" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      { name: "reporter", type: "address" },
      { name: "target", type: "address" },
    ],
    name: "hasReported",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
] as const;

export const POSEIDON_ABI = [
  {
    inputs: [
      { name: "left", type: "bytes32" },
      { name: "right", type: "bytes32" },
    ],
    name: "hash2",
    outputs: [{ name: "", type: "bytes32" }],
    stateMutability: "pure",
    type: "function",
  },
  {
    inputs: [
      { name: "a", type: "bytes32" },
      { name: "b", type: "bytes32" },
      { name: "c", type: "bytes32" },
    ],
    name: "hash3",
    outputs: [{ name: "", type: "bytes32" }],
    stateMutability: "pure",
    type: "function",
  },
] as const;

export const PRIVATE_TRANSFER_ABI = [
  {
    inputs: [{ name: "commitment", type: "bytes32" }],
    name: "shield",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [
      { name: "proof", type: "bytes" },
      { name: "merkleRoot", type: "bytes32" },
      { name: "nullifier", type: "bytes32" },
      { name: "recipient", type: "address" },
      { name: "amount", type: "uint256" },
    ],
    name: "unshield",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "proof", type: "bytes" },
      { name: "merkleRoot", type: "bytes32" },
      { name: "inputNullifiers", type: "bytes32[2]" },
      { name: "outputCommitments", type: "bytes32[2]" },
    ],
    name: "privateTransfer",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [{ name: "root", type: "bytes32" }],
    name: "isKnownRoot",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "nullifier", type: "bytes32" }],
    name: "isSpent",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "getTreeSize",
    outputs: [{ name: "", type: "uint256" }],
    stateMutability: "view",
    type: "function",
  },
] as const;

export const PRIVATE_IDENTITY_ABI = [
  {
    inputs: [{ name: "identityCommitment", type: "bytes32" }],
    name: "registerIdentityCommitment",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      { name: "proof", type: "bytes" },
      { name: "minReputation", type: "uint64" },
      { name: "identityCommitment", type: "bytes32" },
    ],
    name: "proveReputation",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      { name: "proof", type: "bytes" },
      { name: "capabilityHash", type: "bytes32" },
      { name: "identityCommitment", type: "bytes32" },
    ],
    name: "proveCapability",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      { name: "proof", type: "bytes" },
      { name: "minStake", type: "uint256" },
      { name: "identityCommitment", type: "bytes32" },
    ],
    name: "proveStake",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "commitment", type: "bytes32" }],
    name: "isCommitmentRegistered",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
] as const;

export const ZK_VERIFIER_ABI = [
  {
    inputs: [
      { name: "verifyingKeyId", type: "bytes32" },
      { name: "proof", type: "bytes" },
      { name: "publicInputs", type: "uint256[]" },
    ],
    name: "verifyGroth16",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [{ name: "vk", type: "bytes" }],
    name: "registerVerifyingKey",
    outputs: [{ name: "keyId", type: "bytes32" }],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [{ name: "keyId", type: "bytes32" }],
    name: "isKeyRegistered",
    outputs: [{ name: "", type: "bool" }],
    stateMutability: "view",
    type: "function",
  },
] as const;

export const TRUST_BLOCKED = 0;
export const TRUST_UNKNOWN = 1;
export const TRUST_LIMITED = 2;
export const TRUST_FULL = 3;

export type TrustLevel =
  | typeof TRUST_BLOCKED
  | typeof TRUST_UNKNOWN
  | typeof TRUST_LIMITED
  | typeof TRUST_FULL;

const TRUST_LEVEL_NAMES: Record<number, string> = {
  [TRUST_BLOCKED]: "blocked",
  [TRUST_UNKNOWN]: "unknown",
  [TRUST_LIMITED]: "limited",
  [TRUST_FULL]: "full",
};

export function trustLevelName(level: number): string {
  return TRUST_LEVEL_NAMES[level] ?? "unknown";
}
