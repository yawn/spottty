package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/yawn/instagpu/database"
	"github.com/yawn/instagpu/database/filter"
	"github.com/yawn/instagpu/provider"
	"github.com/yawn/instagpu/provider/aws"
)

var showCache bool
var showDatabasePath string
var showFilterMaxResults uint16
var showProviderAWS bool
var showTimeout time.Duration

var showCmd = &cobra.Command{

	Use:   "show",
	Short: "Show a list of candiate instances",
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, _ := context.WithTimeout(context.Background(), showTimeout)

		var providers []provider.Provider

		if showProviderAWS {

			provider, err := aws.New(ctx)

			if err != nil {
				return errors.Wrapf(err, "failed to configure aws")
			}

			providers = append(providers, provider)

		}

		if len(providers) == 0 {
			return fmt.Errorf("no providers selected")
		}

		var filters []filter.Filter

		for _, flag := range filter.Flags {

			if flag.IsSet() {
				filters = append(filters, flag.Apply())
				slog.Debug("filter active", slog.String("name", flag.Name()))
			}

		}

		var (
			db  database.Database
			err error
		)

		if showCache {
			db, err = database.Load(showDatabasePath)
		}

		if !showCache || err != nil {

			db, err = database.New(ctx)

			if err != nil {
				return errors.Wrapf(err, "failed to initialize database")
			}

		}

		results := db.Filter(showFilterMaxResults, filters...)

		for _, result := range results {
			fmt.Println(result)
		}

		return nil

	},
}

func init() {

	flags := showCmd.Flags()

	flags.BoolVar(&showCache, "cache", true, "Enable caching")
	flags.BoolVar(&showProviderAWS, "provider-aws", true, "Enable AWS")
	flags.DurationVar(&showTimeout, "timeout", 30*time.Second, "Timeout for all API operations")
	flags.StringVar(&showDatabasePath, "database-path", "database.json", "Path to the pricing database, for caching")
	flags.Uint16Var(&showFilterMaxResults, "filter-max-results", 10, "Filters by maximum results")

	for _, flag := range filter.Flags {
		flag.Install(flags)
	}

	rootCmd.AddCommand(showCmd)

}
