# Purpose

This project will replace [nodules](https://github.com/unixpickle/nodules). I will use it on my VPS to host my domains and their various worker tasks.

# Language Choice

Last time I wrote a project like this, I chose to use CoffeeScript and Node.js. This time around, I am using Go. I hope this will improve performance, maintainability, and stability.

**Disclaimer**: At the time of starting this project, I have been programming in Go for roughly two days. I bet it's fine, though, because Go is super easy to learn.

# TODO

## Roadmap

Here is a general TODO list which outlines everything that must be done before Goule will be usable.

 * Update [router.go](src/router.go) to forward to the admin site if applicable
   * Setup static file server
 * Create HTTP reverse proxy
 * Apply HTTP proxy to forward rules
 * Create executable management system
   * Some sort of **manager** field for every executable
   * Logging to a file
   * Logging to websockets when requested
   * Relaunch parameters
   * Functions for later use in the API: add/delete/update/stop/start/restart
 * Implement web interface
   * JSON/AJAX APIs
   * Hellish nightmare of implementing a GUI
