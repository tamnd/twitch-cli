package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) streamsCmd() *cobra.Command {
	var game string
	cmd := &cobra.Command{
		Use:   "streams",
		Short: "List top live streams",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := a.effectiveLimit(20)
			streams, err := a.client.Streams(cmd.Context(), game, limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(streams, len(streams))
		},
	}
	cmd.Flags().StringVarP(&game, "game", "g", "", "filter by game/category name")
	return cmd
}

func (a *App) categoriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "categories",
		Short: "List top categories/games by viewer count",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := a.effectiveLimit(20)
			cats, err := a.client.Categories(cmd.Context(), limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(cats, len(cats))
		},
	}
	return cmd
}

func (a *App) searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for channels by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channels, err := a.client.SearchChannels(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(channels, len(channels))
		},
	}
	return cmd
}
