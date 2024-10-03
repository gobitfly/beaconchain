package commands

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/db"

	"github.com/pkg/errors"
)

type AppBundleCommand struct {
	FlagSet *flag.FlagSet
	Config  appBundleCommandConfig
}

type appBundleCommandConfig struct {
	DryRun            bool
	Force             bool // bypass summary confirm
	BundleURL         string
	BundleVersionCode int64
	NativeVersionCode int64
	TargetInstalls    int64
}

func (s *AppBundleCommand) ParseCommandOptions() {
	s.FlagSet.Int64Var(&s.Config.BundleVersionCode, "version-code", 0, "Version code of that bundle (Default: Next)")
	s.FlagSet.Int64Var(&s.Config.NativeVersionCode, "min-native-version", 0, "Minimum required native version (Default: Current)")
	s.FlagSet.Int64Var(&s.Config.TargetInstalls, "target-installs", -1, "How many people to roll out to (Default: All)")
	s.FlagSet.StringVar(&s.Config.BundleURL, "bundle-url", "", "URL to bundle that contains the update, bundle.zip")
	s.FlagSet.BoolVar(&s.Config.Force, "force", false, "Skips summary and confirmation")
}

// nolint
func (s *AppBundleCommand) Run() error {
	if s.Config.BundleURL == "" {
		s.showHelp()
		return errors.New("Please provide a valid bundle URL via --bundle-url")
	}
	if s.Config.BundleVersionCode == 0 {
		fileName := strings.Split(s.Config.BundleURL, "/")
		if len(fileName) == 0 {
			return errors.New("Invalid bundle URL")
		}

		split := strings.Split(fileName[len(fileName)-1], "_")
		if len(split) < 2 {
			return errors.New("Invalid bundle URL")
		}

		// split[1] is the version code
		_, err := fmt.Sscanf(split[1], "%d", &s.Config.BundleVersionCode)
		if err != nil {
			return errors.Wrap(err, "Error parsing version code")
		}
	}
	if s.Config.NativeVersionCode <= 0 {
		err := db.ReaderDb.Get(&s.Config.NativeVersionCode, "SELECT MAX(min_native_version) FROM mobile_app_bundles")
		if err != nil {
			return errors.Wrap(err, "Error getting max native version")
		}
	}

	if s.Config.TargetInstalls < 0 {
		s.Config.TargetInstalls = -1
	}

	if !s.Config.Force {
		// Summary
		fmt.Printf("\n=== Bundle Summary ===")
		fmt.Printf("\nBundle URL: %s", s.Config.BundleURL)
		fmt.Printf("\nBundle Version Code: %d", s.Config.BundleVersionCode)
		fmt.Printf("\nMinimum Native Version: %d", s.Config.NativeVersionCode)
		if s.Config.TargetInstalls == -1 {
			fmt.Printf("\nTarget Installs: All")
		} else {
			fmt.Printf("\nTarget Installs: %d", s.Config.TargetInstalls)
		}
		fmt.Printf("\n======================\n")

		// ask for y/n input
		fmt.Printf("\nDo you want to add this bundle? (y/n)\n")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			return errors.Wrap(err, "Error reading input")
		}

		if input != "y" {
			fmt.Printf("Bundle not added\n")
			return nil
		}
	}

	if s.Config.DryRun {
		fmt.Printf("Dry run, not adding bundle\n")
		return nil
	}

	_, err := db.WriterDb.Exec("INSERT INTO mobile_app_bundles (bundle_url, bundle_version, min_native_version, target_count) VALUES ($1, $2, $3, $4)", s.Config.BundleURL, s.Config.BundleVersionCode, s.Config.NativeVersionCode, s.Config.TargetInstalls)
	if err != nil {
		return errors.Wrap(err, "Error inserting app bundle")
	}

	fmt.Printf("\nBundle added successfully")
	return nil
}

// nolint
func (s *AppBundleCommand) showHelp() {
	fmt.Printf("Usage: app_bundle [options]")
	fmt.Printf("Options:")
	fmt.Printf("  --version-code int\tVersion code of that bundle")
	fmt.Printf("  --min-native-version int\tMinimum required native version (Default: Current)")
	fmt.Printf("  --target-installs int\tHow many people to roll out to (Default: All)")
	fmt.Printf("  --bundle-url string\tURL to bundle that contains the update, bundle.zip")
}
