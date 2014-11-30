# Purpose

This project will replace [nodules](https://github.com/unixpickle/nodules). I will use it on my VPS to host my domains and their various worker tasks.

# Language Choice

Last time I wrote a project like this, I chose to use CoffeeScript and Node.js. This time around, I am using Go. I hope this will improve performance, maintainability, and stability.

**Disclaimer**: At the time of starting this project, I have been programming in Go for roughly two days. I bet it's fine, though, because Go is super easy to learn.

# TODO

## Roadmap

Here is a general TODO list which outlines everything that must be done before Goule will be usable.

 * Simplify the code and use DRY as much as possible.
   * Use reflection for AJAX API calls
   * Make simpler API for hashing passwords
   * Create stubs for the shared locking/saving code in Overseer
   * Use io.LimitedReader for httputil.ReadRequest
 * Create tests for server
 * Create tests for websocket proxy
 * Beef up tests for HTTP reverse proxy
   * Different transfer encodings
   * Various hop-by-hop headers
   * Cookies
 * Implement web interface
   * Finish admin page
     * Create editor for TLS configuration
   * Create "add service" page
   * Slim down jquery-ui configuration
   * Resize images
   * Compile CSS files
   * Compile JS files
   * Write the rest of the AJAX API stubs
     * Add service
     * Rename service
     * Set TLS
     * Set service rules
     * Set service execs
 * Logging to a file
 * Buffered/truncated logging
 * Set CAs in HTTPs server

## Done

I'm moving things from my TODO list to this spot once I do them:

 * Update router.go to forward to the admin site if applicable
   * Begin JSON/AJAX control API
   * Setup static file server
 * Make admin HTTP static file server secure!
 * Create executable management system
    * Targetted start/stop functions (no restart, for various reasons)
    * Functionality: automatic relaunch, stop, start, status
 * Rewrite executable system to be nicer (once I know how)
 * Create tests for StoppableLock
 * Add errors and launch/exit dates to ExecutableInfo
 * Begin JSON/AJAX APIs for managing services and executables
   * Make Sessions a part of Overseer
   * Create request context for routing handlers
   * Implement login AJAX call
   * Call for listing services
   * Call for changing the password
   * Call for changing HTTP settings
   * Call for changing HTTPS settings
   * Call for changing TLS settings
   * Call for changing admin forward rules
   * Call for changing service's name
   * Call for changing service's forward rules
   * Call for changing service's executables
 * Restructured entire thing to use subpackages
 * Create deep-copy methods for all internal data structures
 * Use deep-copy methods for new Overseer
 * Restructured for `go get`
 * Implement web interface
   * Login UI
 * Create HTTP reverse proxy
 * Create tests for HTTP reverse proxy
 * Apply HTTP proxy to forward rules
 * Create tests for config
 * Implement more APIs
   * Get full configuration
   * Set admin session timeout
   * Set TLS
   * Add service
   * Set proxy settings
 * Implement web interface
   * Finish admin page
     * Implement HTTP/HTTPS settings
     * Implement session timeout setting
     * Implement proxy settings
     * Line up fields and make them look nicer
     * Create list of forward rules
   * Write the rest of the AJAX API stubs
     * Set admin rules

## Possible TODOs down the road

 * Logging to websockets
