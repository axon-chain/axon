package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Agent module transactions",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		CmdRegister(),
		CmdAddStake(),
		CmdDeregister(),
		CmdHeartbeat(),
		CmdUpdateAgent(),
		CmdSubmitAIChallenge(),
		CmdRevealAIChallenge(),
	)

	return txCmd
}

func CmdRegister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [capabilities] [model] [stake]",
		Short: "Register as an AI agent",
		Long:  "Register as an AI agent on the Axon network.\nExample: axond tx agent register \"coding,reasoning\" \"gpt-4\" 100000000000000000000aaxon",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			stake, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid stake amount: %w", err)
			}

			msg := &types.MsgRegister{
				Sender:       clientCtx.GetFromAddress().String(),
				Capabilities: args[0],
				Model:        args[1],
				Stake:        stake,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdAddStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-stake [amount]",
		Short: "Add stake to an existing AI agent",
		Long:  "Increase the stake of an already-registered AI agent.\nExample: axond tx agent add-stake 50000000000000000000aaxon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			stake, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid stake amount: %w", err)
			}

			msg := &types.MsgAddStake{
				Sender: clientCtx.GetFromAddress().String(),
				Stake:  stake,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDeregister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Request to deregister as an AI agent (7-day cooldown)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgDeregister{
				Sender: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdHeartbeat() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "heartbeat",
		Short: "Send a heartbeat to maintain online status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgHeartbeat{
				Sender: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [capabilities] [model]",
		Short: "Update agent capabilities and model",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateAgent{
				Sender:       clientCtx.GetFromAddress().String(),
				Capabilities: args[0],
				Model:        args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSubmitAIChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-challenge [epoch] [answer]",
		Short: "Submit a commit hash for an AI challenge (commit phase)",
		Long:  "Submits sha256(address+\":\"+answer) as the commit hash. The sender address is used as implicit salt. You must reveal the answer later.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch: %w", err)
			}

			commitInput := clientCtx.GetFromAddress().String() + ":" + args[1]
			hash := sha256.Sum256([]byte(commitInput))
			commitHash := hex.EncodeToString(hash[:])

			msg := &types.MsgSubmitAIChallengeResponse{
				Sender:     clientCtx.GetFromAddress().String(),
				Epoch:      epoch,
				CommitHash: commitHash,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRevealAIChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveal-challenge [epoch] [answer]",
		Short: "Reveal the answer for a previously committed AI challenge",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch: %w", err)
			}

			msg := &types.MsgRevealAIChallengeResponse{
				Sender:     clientCtx.GetFromAddress().String(),
				Epoch:      epoch,
				RevealData: args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
