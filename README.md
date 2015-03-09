# Purpose

This project will replace [nodules](https://github.com/unixpickle/nodules). I will use it on my VPS to host my domains and their various worker tasks.

# License

**goule** is licensed under the BSD 2-clause license. See [LICENSE](LICENSE).

```
Copyright (c) 2015, Alex Nichol.
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer. 
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

# TODO

## Roadmap

Here is a general TODO list which outlines everything that must be done before Goule will be usable.

 * Re-organize session management
 * Implement web interface
   * Simple AJAX/JavaScript API
   * Login page
   * Server settings page
   * Rules page
   * Services page

## Done

I'm moving things from my TODO list to this spot once I do them:

 * Start over using external APIs for all functionality.
 * Move Service type to executor library
 * Create configuration structure
 * Create default configuration generator
 * Create HTTP endpoints
 * Create API handler mechanism
 * Create service routers
 * Implement APIs
   * Set admin port
   * Set admin session timeout
   * Set admin assets path
   * Edit TLS info
   * Delete service
   * Add service
   * Add rule
   * Replace rule
   * Delete rule
   * Modify service
   * Start service
   * Stop service
   * Getters for config and services
   * Set HTTP and HTTPS configuration
   * Unify API errors
 * Implement web interface
   * AJAX session management
