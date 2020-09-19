package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexzorin/authy"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var (
	countrycode, mobile, password string

	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "Authy account info or register device",
		Long: `Register device or show registered account info. 

Can specify country code, mobile number and authy main password.
If not provided, will get from command line stdin`,
		Run: func(cmd *cobra.Command, args []string) {
			registerOrGetDeviceInfo()
		},
	}
)

const (
	configFileName = ".authy.json"
)

type deviceRegistration struct {
	UserID       uint64 `json:"user_id,omitempty"`
	DeviceID     uint64 `json:"device_id,omitempty"`
	Seed         string `json:"seed,omitempty"`
	APIKey       string `json:"api_key,omitempty"`
	MainPassword string `json:"main_password,omitempty"`
}

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.Flags().StringVarP(&countrycode, "countrycode", "c", "", "phone number country code (e.g. 1 for United States), digitals only")
	accountCmd.Flags().StringVarP(&mobile, "mobilenumber", "m", "", "phone number, digitals only")
	accountCmd.Flags().StringVarP(&password, "password", "p", "", "authy main password")
}

func registerOrGetDeviceInfo() {
	devInfo, err := LoadExistingDeviceInfo()
	if err == nil {
		log.Println("device info found")
		log.Printf("%+v\n", devInfo)
		return
	}

	if os.IsNotExist(err) {
		devInfo, err = newRegistrationDevice()
		if err != nil {
			os.Exit(1)
		}

		log.Println("Register device success!!!")
		log.Printf("Your device info: %+v\n", devInfo)
		os.Exit(0)
	}

	if err != nil {
		log.Println("Load device info failed", err)
	}
}

func newRegistrationDevice() (devInfo deviceRegistration, err error) {
	var (
		sc      = bufio.NewScanner(os.Stdin)
		phoneCC int
	)

	if len(countrycode) == 0 {
		fmt.Print("\nWhat is your phone number's country code? (digits only, e.g. 86): ")
		if !sc.Scan() {
			err = errors.New("Please provide a phone country code, e.g. 86")
			log.Println(err)
			return
		}

		countrycode = sc.Text()
	}

	phoneCC, err = strconv.Atoi(strings.TrimSpace(countrycode))
	if err != nil {
		log.Println("Invalid country code. Parse country code failed", err)
		return
	}

	if len(mobile) == 0 {
		fmt.Print("\nWhat is your phone number? (digits only): ")
		if !sc.Scan() {
			err = errors.New("Please provide a phone number, e.g. 1232211")
			log.Println(err)
			return
		}

		mobile = sc.Text()
	}

	mobile = strings.TrimSpace(mobile)

	client, err := authy.NewClient()
	if err != nil {
		log.Println("New authy client failed", err)
		return
	}

	userStatus, err := client.QueryUser(nil, phoneCC, mobile)
	if err != nil {
		log.Println("Query user failed", err)
		return
	}

	if !userStatus.IsActiveUser() {
		err = errors.New("There doesn't seem to be an Authy account attached to that phone number")
		log.Println(err)
		return
	}

	// Begin a device registration using Authy app push notification
	regStart, err := client.RequestDeviceRegistration(nil, userStatus.AuthyID, authy.ViaMethodPush)
	if err != nil {
		log.Println("Start register device failed", err)
		return
	}

	if !regStart.Success {
		err = fmt.Errorf("Authy did not accept the device registration request: %+v", regStart)
		log.Println(err)
		return
	}

	var regPIN string
	timeout := time.Now().Add(5 * time.Minute)
	for {
		if timeout.Before(time.Now()) {
			err = errors.New("Gave up waiting for user to respond to Authy device registration request")
			log.Println(err)
			return
		}

		log.Printf("Checking device registration status (%s until we give up)", time.Until(timeout).Truncate(time.Second))

		regStatus, err1 := client.CheckDeviceRegistration(nil, userStatus.AuthyID, regStart.RequestID)
		if err1 != nil {
			err = err1
			log.Println(err)
			return
		}
		if regStatus.Status == "accepted" {
			regPIN = regStatus.PIN
			break
		} else if regStatus.Status != "pending" {
			err = fmt.Errorf("Invalid status while waiting for device registration: %s", regStatus.Status)
			log.Println(err)
			return
		}

		time.Sleep(5 * time.Second)
	}

	regComplete, err := client.CompleteDeviceRegistration(nil, userStatus.AuthyID, regPIN)
	if err != nil {
		log.Println(err)
		return
	}

	if regComplete.Device.SecretSeed == "" {
		err = errors.New("Something went wrong completing the device registration")
		log.Println(err)
		return
	}

	devInfo = deviceRegistration{
		UserID:   regComplete.AuthyID,
		DeviceID: regComplete.Device.ID,
		Seed:     regComplete.Device.SecretSeed,
		APIKey:   regComplete.Device.APIKey,
	}

	err = SaveDeviceInfo(devInfo)
	if err != nil {
		log.Println("Save device info failed", err)
	}

	return
}

// SaveDeviceInfo ..
func SaveDeviceInfo(devInfo deviceRegistration) (err error) {
	regrPath, err := ConfigPath(configFileName)
	if err != nil {
		return
	}

	f, err := os.OpenFile(regrPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	defer f.Close()
	err = json.NewEncoder(f).Encode(devInfo)
	return
}

// LoadExistingDeviceInfo ,,,
func LoadExistingDeviceInfo() (devInfo deviceRegistration, err error) {
	devPath, err := ConfigPath(configFileName)
	if err != nil {
		log.Println("Get device info file path failed", err)
		os.Exit(1)
	}

	f, err := os.Open(devPath)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&devInfo)
	return
}

// ConfigPath get config file path
func ConfigPath(fname string) (string, error) {
	devPath, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(devPath, fname), nil
}
