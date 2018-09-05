// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"time"

	"github.com/spf13/viper"

	"github.com/briandowns/spinner"
	"github.com/gobuffalo/pop"
	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v1"
	gomail "gopkg.in/gomail.v2"

	"github.com/muacm/courier/models"
)

type Context struct {
	Date    string
	Contact models.Contact
}

// sendCmd represents the list command
var sendCmd = &cobra.Command{
	Use:   "send [MEETING DATE] [TEMPLATE]",
	Short: "A brief description of your command",
	Long: `send: Sends a formatted email to the list of people in the sqlite db
	
	[MEETING DATE] -> Date in the format MM-DD-YYYY
	[TEMPLATE]     -> HTML Template for the email
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dry, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			panic(err)
		}

		db, err := pop.Connect("production")
		if err != nil {
			panic(err)
		}

		contacts := models.Contacts{}

		only, err := cmd.Flags().GetString("only")
		if err != nil {
			panic(err)
		}

		if only == "" {
			err = db.All(&contacts)
		} else {
			err = db.Where("name LIKE ?", only).All(&contacts)
		}

		if err != nil {
			panic(err)
		}

		// parse the args
		date := args[0]
		file := args[1]

		format, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}

		tmpl, err := template.New("test").Parse(string(format))
		if err != nil {
			panic(err)
		}

		ctx := Context{
			Date: date,
		}

		passw := viper.GetString("outlook.password")
		username := viper.GetString("outlook.username")

		sp := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		sp.Suffix = emoji.Sprint("\t:email:\tDialing Outlook!")
		sp.Start()
		d := gomail.NewDialer("smtp.office365.com", 587, username, passw)
		s, err := d.Dial()
		if err != nil {
			panic(err)
		}
		sp.Stop()
		fmt.Println(emoji.Sprint("\t:heavy_check_mark:\tDialing Success!"))

		bar := pb.StartNew(len(contacts))
		for _, contact := range contacts {
			ctx.Contact = contact

			var buffer bytes.Buffer
			err = tmpl.Execute(&buffer, ctx)
			if err != nil {
				panic(err)
			}

			m := gomail.NewMessage()
			m.SetHeader("From", username)
			m.SetAddressHeader("To", ctx.Contact.Email, ctx.Contact.Name)
			m.SetHeader("Subject", fmt.Sprintf("ACM/UPE Meeting - %s", ctx.Date))
			m.SetBody("text/html", buffer.String())

			if !dry {
				if err := gomail.Send(s, m); err != nil {
					panic(err)
				}
			} else {
				time.Sleep(2 * time.Second)
			}

			fmt.Println(emoji.Sprintf("\r\033[K\t:heavy_check_mark:\tSent: %s", contact.Name))
			bar.Increment()
		}

		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	sendCmd.Flags().Bool("dry-run", false, "Whether or not to send the emails.")
	sendCmd.Flags().String("only", "", "Send only to this group")
}
