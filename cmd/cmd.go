package cmd

import "sensor-service/cmd/api"

// Execute command
func Execute() {
	//var rootCmd = &cobra.Command{
	//	Use:   "help",
	//	Short: "Service command list",
	//	Long:  "Helping get service command list",
	//}
	//
	//commands := []*cobra.Command{
	//	{
	//		Use:   "serve",
	//		Short: "Listening HTTP request",
	//		Long:  "Listening HTTP request",
	//		Run: func(cmd *cobra.Command, args []string) {
	//			_, err := api.NewServer()
	//			if err != nil {
	//				panic(err)
	//			}
	//		},
	//	},
	//}
	//
	//for _, command := range commands {
	//	rootCmd.AddCommand(command)
	//}
	//
	//err := rootCmd.Execute()
	//if err != nil {
	//	panic(err)
	//}
	_, err := api.NewServer()
	if err != nil {
		panic(err)
	}

}
