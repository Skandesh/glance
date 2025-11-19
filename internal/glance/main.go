package glance

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var buildVersion = "dev"

// validateProductionEnvironment checks critical environment variables and configuration
func validateProductionEnvironment() {
	warnings := []string{}
	errors := []string{}

	// Check GLANCE_MASTER_KEY
	masterKey := os.Getenv("GLANCE_MASTER_KEY")
	if masterKey == "" {
		warnings = append(warnings, "GLANCE_MASTER_KEY not set - using insecure default key. SET THIS IN PRODUCTION!")
	} else if len(masterKey) < 32 {
		warnings = append(warnings, fmt.Sprintf("GLANCE_MASTER_KEY is too short (%d chars). Recommended: 32+ characters", len(masterKey)))
	}

	// Check if Stripe keys are configured
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey != "" {
		// Validate Stripe key format
		if !strings.HasPrefix(stripeKey, "sk_") {
			errors = append(errors, "STRIPE_SECRET_KEY must start with 'sk_' prefix")
		}

		// Check if using test mode in what appears to be production
		if strings.HasPrefix(stripeKey, "sk_test_") {
			warnings = append(warnings, "Using Stripe TEST mode key. Switch to 'sk_live_' for production")
		} else if strings.HasPrefix(stripeKey, "sk_live_") {
			slog.Info("Stripe LIVE mode detected - production configuration")
		}
	}

	// Check webhook secret
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if stripeKey != "" && webhookSecret == "" {
		warnings = append(warnings, "STRIPE_WEBHOOK_SECRET not set - real-time updates will NOT work")
	}

	// Print errors (fatal)
	if len(errors) > 0 {
		fmt.Println("\n❌ CRITICAL CONFIGURATION ERRORS:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println("\nFix these errors before starting the application.")
		os.Exit(1)
	}

	// Print warnings (non-fatal but important)
	if len(warnings) > 0 {
		fmt.Println("\n⚠️  CONFIGURATION WARNINGS:")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()
	}

	// Production readiness check
	if masterKey != "" && len(masterKey) >= 32 && strings.HasPrefix(stripeKey, "sk_live_") && webhookSecret != "" {
		slog.Info("✅ Production environment validation passed")
	}
}

func Main() int {
	options, err := parseCliOptions()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	switch options.intent {
	case cliIntentVersionPrint:
		fmt.Println(buildVersion)
	case cliIntentServe:
		// remove in v0.10.0
		if serveUpdateNoticeIfConfigLocationNotMigrated(options.configPath) {
			return 1
		}

		if err := serveApp(options.configPath); err != nil {
			fmt.Println(err)
			return 1
		}
	case cliIntentConfigValidate:
		contents, _, err := parseYAMLIncludes(options.configPath)
		if err != nil {
			fmt.Printf("Could not parse config file: %v\n", err)
			return 1
		}

		if _, err := newConfigFromYAML(contents); err != nil {
			fmt.Printf("Config file is invalid: %v\n", err)
			return 1
		}
	case cliIntentConfigPrint:
		contents, _, err := parseYAMLIncludes(options.configPath)
		if err != nil {
			fmt.Printf("Could not parse config file: %v\n", err)
			return 1
		}

		fmt.Println(string(contents))
	case cliIntentSensorsPrint:
		return cliSensorsPrint()
	case cliIntentMountpointInfo:
		return cliMountpointInfo(options.args[1])
	case cliIntentDiagnose:
		runDiagnostic()
	case cliIntentSecretMake:
		key, err := makeAuthSecretKey(AUTH_SECRET_KEY_LENGTH)
		if err != nil {
			fmt.Printf("Failed to make secret key: %v\n", err)
			return 1
		}

		fmt.Println(key)
	case cliIntentPasswordHash:
		password := options.args[1]

		if password == "" {
			fmt.Println("Password cannot be empty")
			return 1
		}

		if len(password) < 6 {
			fmt.Println("Password must be at least 6 characters long")
			return 1
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Failed to hash password: %v\n", err)
			return 1
		}

		fmt.Println(string(hashedPassword))
	}

	return 0
}

func serveApp(configPath string) error {
	// Validate production environment before starting
	validateProductionEnvironment()

	// TODO: refactor if this gets any more complex, the current implementation is
	// difficult to reason about due to all of the callbacks and simultaneous operations,
	// use a single goroutine and a channel to initiate synchronous changes to the server
	exitChannel := make(chan struct{})
	hadValidConfigOnStartup := false
	var stopServer func() error

	onChange := func(newContents []byte) {
		if stopServer != nil {
			log.Println("Config file changed, reloading...")
		}

		config, err := newConfigFromYAML(newContents)
		if err != nil {
			log.Printf("Config has errors: %v", err)

			if !hadValidConfigOnStartup {
				close(exitChannel)
			}

			return
		}

		app, err := newApplication(config)
		if err != nil {
			log.Printf("Failed to create application: %v", err)

			if !hadValidConfigOnStartup {
				close(exitChannel)
			}

			return
		}

		if !hadValidConfigOnStartup {
			hadValidConfigOnStartup = true
		}

		if stopServer != nil {
			if err := stopServer(); err != nil {
				log.Printf("Error while trying to stop server: %v", err)
			}
		}

		go func() {
			var startServer func() error
			startServer, stopServer = app.server()

			if err := startServer(); err != nil {
				log.Printf("Failed to start server: %v", err)
			}
		}()
	}

	onErr := func(err error) {
		log.Printf("Error watching config files: %v", err)
	}

	configContents, configIncludes, err := parseYAMLIncludes(configPath)
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	stopWatching, err := configFilesWatcher(configPath, configContents, configIncludes, onChange, onErr)
	if err == nil {
		defer stopWatching()
	} else {
		log.Printf("Error starting file watcher, config file changes will require a manual restart. (%v)", err)

		config, err := newConfigFromYAML(configContents)
		if err != nil {
			return fmt.Errorf("validating config file: %w", err)
		}

		app, err := newApplication(config)
		if err != nil {
			return fmt.Errorf("creating application: %w", err)
		}

		startServer, _ := app.server()
		if err := startServer(); err != nil {
			return fmt.Errorf("starting server: %w", err)
		}
	}

	<-exitChannel
	return nil
}

func serveUpdateNoticeIfConfigLocationNotMigrated(configPath string) bool {
	if !isRunningInsideDockerContainer() {
		return false
	}

	if _, err := os.Stat(configPath); err == nil {
		return false
	}

	// glance.yml wasn't mounted to begin with or was incorrectly mounted as a directory
	if stat, err := os.Stat("glance.yml"); err != nil || stat.IsDir() {
		return false
	}

	templateFile, _ := templateFS.Open("v0.7-update-notice-page.html")
	bodyContents, _ := io.ReadAll(templateFile)

	fmt.Println("!!! WARNING !!!")
	fmt.Println("The default location of glance.yml in the Docker image has changed starting from v0.7.0.")
	fmt.Println("Please see https://github.com/glanceapp/glance/blob/main/docs/v0.7.0-upgrade.md for more information.")

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(bodyContents))
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()

	return true
}
