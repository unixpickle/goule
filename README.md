# Purpose

This project will replace [nodules](https://github.com/unixpickle/nodules). I will use it on my VPS to host my domains and their various worker tasks.

# Language Choice

Last time I wrote a project like this, I chose to use CoffeeScript and Node.js. This time around, I am using Go. I hope this will improve performance, maintainability, and stability.

**Disclaimer**: At the time of starting this project, I have been programming in Go for roughly two days. I bet it's fine, though, because Go is super easy to learn.

# TODO

## Roadmap

Here is a general TODO list which outlines everything that must be done before Goule will be usable.

 * Add features to executable system
   * Relaunching + extended state: IsRunning, IsRelaunching
   * Targetted start/stop/restart functions
   * Logging to a file
 * Create HTTP reverse proxy
 * Apply HTTP proxy to forward rules
 * Implement web interface
   * JSON/AJAX APIs
   * Hellish nightmare of implementing a GUI

## Done

I'm moving things from my TODO list to this spot once I do them:

 * Update [router.go](src/router.go) to forward to the admin site if applicable
   * Begin JSON/AJAX control API
   * Setup static file server
 * Make admin HTTP static file server secure!
 * Create executable management system

## Possible features down the road

 * Logging to websockets
 * Logging which removes old data to avoid overflow
