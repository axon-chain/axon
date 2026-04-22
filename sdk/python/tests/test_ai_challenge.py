"""Tests for AI Challenge client."""

import pytest
from unittest.mock import MagicMock, patch
from axon.ai_challenge import AIChallengeClient


class TestAIChallengeClient:
    """Tests for AIChallengeClient."""
    
    def test_normalize_answer(self):
        """Test answer normalization."""
        # Basic lowercase
        assert AIChallengeClient.normalize_answer("HELLO") == "hello"
        
        # Remove spaces
        assert AIChallengeClient.normalize_answer("hello world") == "helloworld"
        
        # Remove tabs and newlines
        assert AIChallengeClient.normalize_answer("hello\nworld") == "helloworld"
        
        # Mixed case with punctuation
        assert AIChallengeClient.normalize_answer("  Hello  World! ") == "helloworld!"
    
    def test_hash_answer(self):
        """Test answer hashing."""
        # Should produce consistent hash
        answer = "42"
        hash1 = AIChallengeClient.hash_answer(answer)
        hash2 = AIChallengeClient.hash_answer(answer)
        assert hash1 == hash2
    
    @pytest.fixture
    def mock_client(self):
        """Create mock client for testing."""
        with patch('axon.ai_challenge.Web3') as mock_w3:
            mock_w3_instance = MagicMock()
            mock_w3.HTTPProvider.return_value = mock_w3_instance
            mock_w3_instance.is_connected.return_value = True
            mock_w3.eth = MagicMock()
            mock_w3.eth.chain_id = 8210
            mock_w3.eth.block_number = 1000
            mock_w3.keccak = MagicMock(return_value=b'\x00' * 32)
            
            client = AIChallengeClient("https://mainnet-rpc.axonchain.ai/")
            return client
    
    def test_get_active_challenge(self, mock_client):
        """Test getting active challenge."""
        challenge = mock_client.get_active_challenge()
        
        assert 'challenge_id' in challenge
        assert 'question' in challenge
        assert 'submit_deadline' in challenge
        assert 'reveal_deadline' in challenge
    
    def test_get_challenge_result(self, mock_client):
        """Test getting challenge result."""
        result = mock_client.get_challenge_result(1)
        
        assert 'correct_answer' in result
        assert 'participants' in result
        assert 'rewards_distributed' in result


if __name__ == "__main__":
    pytest.main([__file__, "-v"])