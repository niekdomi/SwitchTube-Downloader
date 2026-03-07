package cmd

import (
	"strings"

	"switchtube-downloader/internal/download"
	"switchtube-downloader/internal/models"

	"github.com/spf13/cobra"
)


// init initializes the download command and adds it to the root command with its flags.
func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolP("episode", "e", false, "Prefixes the video with episode-number e.g. 01_OR_Mapping.mp4")
	downloadCmd.Flags().BoolP("skip", "s", false, "Skip video if it already exists")
	downloadCmd.Flags().BoolP("force", "f", false, "Force overwrite if file already exist")
	downloadCmd.Flags().BoolP("all", "a", false, "Download the whole content of a channel")
	downloadCmd.Flags().StringP("output", "o", "", "Output directory for downloaded files")
}

var downloadCmd = &cobra.Command{
	Use:   "download <id|url> [id|url]...",
	Short: "Download one or more videos or channels",
	Long: "Download one or more videos or channels. Automatically detects for each input whether it is a video or channel.\n" +
		"You can also pass the whole URL instead of the ID for convenience.",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		episode, err := cmd.Flags().GetBool("episode")
		if err != nil {
			log.Error("Error getting episode flag", "err", err)

			return
		}

		skip, err := cmd.Flags().GetBool("skip")
		if err != nil {
			log.Error("Error getting skip flag", "err", err)

			return
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Error("Error getting force flag", "err", err)

			return
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Error("Error getting all flag", "err", err)

			return
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Error("Error getting output flag", "err", err)

			return
		}

		for _, arg := range args {
			config := models.DownloadConfig{
				Media:      arg,
				UseEpisode: episode,
				Skip:       skip,
				Force:      force,
				All:        all,
				OutputDir:  strings.TrimSpace(output),
			}

			err = download.Download(config)
			if err != nil {
				log.Error("Download failed", "err", err)
			}
		}
	},
}
