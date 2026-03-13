package types

import "fmt"

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Agents: []Agent{},
	}
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, agent := range gs.Agents {
		if agent.Address == "" {
			return fmt.Errorf("agent address cannot be empty")
		}
		if seen[agent.Address] {
			return fmt.Errorf("duplicate agent address: %s", agent.Address)
		}
		seen[agent.Address] = true
		if agent.Reputation > gs.Params.MaxReputation {
			return fmt.Errorf("agent %s reputation %d exceeds max %d", agent.Address, agent.Reputation, gs.Params.MaxReputation)
		}
	}
	return nil
}
