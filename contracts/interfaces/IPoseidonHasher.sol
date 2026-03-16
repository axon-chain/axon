// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

interface IPoseidonHasher {
    function hash2(bytes32 left, bytes32 right) external pure returns (bytes32);
    function hash3(bytes32 a, bytes32 b, bytes32 c) external pure returns (bytes32);
}
