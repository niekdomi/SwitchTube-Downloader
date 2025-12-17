package cmd

import (
	"fmt"
	"strings"

	"switchtube-downloader/internal/download"
	"switchtube-downloader/internal/models"

	"github.com/spf13/cobra"
)

// init initializes the download command and adds it to the root command with
// its flags.
func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().
		BoolP("episode", "e", false, "Prefixes the video with episode-number e.g. 01_OR_Mapping.mp4")
	downloadCmd.Flags().BoolP("skip", "s", false, "Skip video if it already exists")
	downloadCmd.Flags().BoolP("force", "f", false, "Force overwrite if file already exist")
	downloadCmd.Flags().BoolP("all", "a", false, "Download the whole content of a channel")
	downloadCmd.Flags().StringP("output", "o", "", "Output directory for downloaded files")
}

var downloadCmd = &cobra.Command{
	Use:   "download <id|url>",
	Short: "Download a video or channel",
	Long: "Download a video or channel. Automatically detects if input is a video or channel.\n" +
		"You can also pass the whole URL instead of the ID for convenience.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		episode, err := cmd.Flags().GetBool("episode")
		if err != nil {
			fmt.Printf("Error getting episode flag: %v", err)

			return
		}

		skip, err := cmd.Flags().GetBool("skip")
		if err != nil {
			fmt.Printf("Error getting skip flag: %v", err)

			return
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			fmt.Printf("Error getting force flag: %v", err)

			return
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			fmt.Printf("Error getting all flag: %v", err)

			return
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			fmt.Printf("Error getting output flag: %v", err)

			return
		}

		config := models.DownloadConfig{
			Media:      args[0],
			UseEpisode: episode,
			Skip:       skip,
			Force:      force,
			All:        all,
			Output:     strings.TrimSpace(output),
		}

		err = download.Download(config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)

			return
		}
	},
}
