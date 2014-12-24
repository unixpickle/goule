# Purpose

This project will replace [nodules](https://github.com/unixpickle/nodules). I will use it on my VPS to host my domains and their various worker tasks.

# Language Choice

Last time I wrote a project like this, I chose to use CoffeeScript and Node.js. This time around, I am using Go. I hope this will improve performance, maintainability, and stability.

**Disclaimer**: At the time of starting this project, I have been programming in Go for roughly two days. I bet it's fine, though, because Go is super easy to learn.

# TODO

## Roadmap

Here is a general TODO list which outlines everything that must be done before Goule will be usable.

 * Move Service type to executor library
 * Create configuration structure
 * Create default configuration generator
 * Create HTTP endpoints
 * Create API handler mechanism
 * Create service routers
 * Implement APIs
 * Implement web interface

## Done

I'm moving things from my TODO list to this spot once I do them:

 * Start over using external APIs for all functionality.
