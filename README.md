# dotenv

[![PkgGoDev](https://pkg.go.dev/badge/github.com/sbowman/dotenv)](https://pkg.go.dev/github.com/sbowman/dotenv)

The `dotenv` package manages the configuration of a Go application through 
environment variables, similar to the Ruby dotenv library.  It provides helper
functions to pull typed values from environment variables, configure default
values, and support environmental overrides during development through a `.env`
file.

## Getting values

The `dotenv` package supports most of the basic Go data types.  Use one of the
`Get` calls to pull in the value from the system environment variables 
(`os.LookupEnv`), or use the registered default value (see below).

Here's some samples:

    maxConns := dotenv.GetInt("MAX_CONNS")
    timeout := dotenv.GetDuration("HTTP_TIMEOUT")
    hostname := dotenv.GetString("HOSTNAME")
    useTLS := dotenv.GetBool("ENABLE_TLS")
    
_Note: recommend using constants for environment variable names!_

See the Godocs for all available `Get` functions.

These functions are safe to call in multithreaded code, i.e. goroutines.

## Default values

It's frequently useful to have default values for application settings.  For
example, most people can just use the default maximun number of database 
connections, or HTTP timeout values.  The `dotenv` package allows you to 
register these defaults, and if the environment variable hasn't been set, a
call to a `Get` function will return these defaults.

Here's some example registration calls:

    func init() {
        dotenv.Register("VERBOSITY", -1, "Set the debug logging verbosity")
        dotenv.Register("DB_DRIVER", "postgres", "Name of the database driver")
        dotenv.Register("DB_URI", "postgres://postgres@localhost/mydb?sslmode=disable", "Database connection URI")
    }

Provide the environment variable to look for, it's default value, and a 
description of this setting.  

You can also use this to display help information to users.  In your startup
command, if a required setting is missing or incorrect, or maybe the user 
starts things with a `--help` CLI parameter, you may call the `Help()` function
to display the registered settings, their default values, types, and description.

## The .env file

The `dotenv` package also supports a `.env` file.  This file can exist in either
the project directory (or wherever you start the application), or the user's
`$HOME` directory.  The `dotenv` pacakge looks for these files and overrides 
any environment variables or defaults when they exist in one of these files.

The `.env` file just looks like any `.bashrc` or similar environment 
configuration file:

    # Logging
    VERBOSITY=3
    
    # Database settings
    DB_URI=postgres://postgres@localhost/mydb?sslmode=disable
    DB_MIN=2
    DB_MAX=6
    
It's best to add `.env` to the `.gitignore` file in your project, so a local
developer's environment variables--particularly usernames and passwords--don't
accidentally get pushed into source control.

Note:  there's no way for `dotenv` to distinguish between an environment 
variable that's been set in a `.bashrc` or through an `export` in your terminal
session, and one that's been set on the command line, e.g. `VERBOSITY=2 ./myapp`.
Thus, any setting in the `.env` file will override environment settings.  If
you want to customize the setting from the command line, make sure to comment
it out or remove it from the `.env` file.  
