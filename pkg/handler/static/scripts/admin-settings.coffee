window.goule = {} if not window.goule?

class AdminSettings
  constructor: ->
    @serverSettings = new ServerSettings()
    @passwordChanger = new PasswordChanger()
    @forwardRules = new ForwardRules()
    $('#as-container input').on 'input', => @inputChanged()
    $('#as-container input').change => @inputChanged()
  
  show: (animate) ->
    @serverSettings.disable()
    @forwardRules.disable()
    window.goule.api.getConfig (err, config) =>
      if not err?
        @serverSettings.update config
        @serverSettings.enable()
        @forwardRules.update config
        @forwardRules.enable()
      else
        console.log 'failed to get configuration'
    
    container = $ '#as-container'
    if animate
      container.css display: 'inline-block', opacity: '0.0'
      container.animate opacity: 1.0
    else
      container.css display: 'inline-block', opacity: '1.0'
  
  hide: (animate) ->
    container = $ '#as-container'
    if animate
      container = $('#as-container')
      $('#as-container').fadeOut()
    else
      container.css display: 'none'
  
  inputChanged: -> @serverSettings.inputChanged()

class ServerSettings
  constructor: ->
    @config = null
    @saveButton = $ '#as-server-settings .save-button'
    @protoPortInputs =
      http: $ '#as-http-port'
      https: $ '#as-https-port'
    @protoEnableInputs =
      http: $ '#as-http-enabled'
      https: $ '#as-https-enabled'
    @timeoutInput = $ '#as-session-timeout'
    @rewriteInput = $ '#as-rewrite-host'
    @websocketsInput = $ '#as-websockets'
    @saveButton.click => @save()
    @all = $ '#as-server-settings input, #as-server-settings .save-button'
  
  enable: -> @all.css opacity: '1.0', 'pointer-events': 'auto'
  
  disable: -> @all.css opacity: '0.5', 'pointer-events': 'none'
  
  save: ->
    # Create a list of calls to make in order to save the various fields.
    saveCalls = []
    for proto in ['http', 'https']
      do (proto) =>
        if @didProtoChange proto
          saveCalls.push (cb) => @saveProto proto, cb
    if @didTimeoutChange
      saveCalls.push (cb) => @saveTimeout cb
    if @didProxyChange
      saveCalls.push (cb) => @saveProxy cb
    # Run the calls and disable the inputs in the meantime
    remaining = saveCalls.length
    return if remaining is 0
    @disable()
    for aCall in saveCalls
      aCall =>
        return if --remaining isnt 0
        @enable()
        @inputChanged()
  
  inputChanged: ->
    if @didChange()
      @saveButton.css 'display': 'inline-block'
    else
      @saveButton.css 'display': 'none'
  
  update: (config) ->
    @config = config
    @protoPortInputs.http.val '' + config.http.port
    @protoPortInputs.https.val '' + config.https.port
    @protoEnableInputs.http.prop 'checked', config.http.enabled
    @protoEnableInputs.https.prop 'checked', config.https.enabled
    @timeoutInput.val '' + config.admin.session_timeout
    @rewriteInput.prop 'checked', config.proxy.rewrite_host
    @websocketsInput.prop 'checked', config.proxy.websockets
    @inputChanged()
  
  didChange: ->
    return @didProtoChange('http') or @didProtoChange('https') or
      @didTimeoutChange() or @didProxyChange()
  
  getProto: (proto) ->
    enabled = @protoEnableInputs[proto].prop 'checked'
    port = parseInt @protoPortInputs[proto].val()
    port = {'http': 80, 'https': 443}[proto] if isNaN port
    return enabled: enabled, port: port
  
  didProtoChange: (proto) ->
    s = @getProto proto
    cfg = @config[proto]
    return s.port isnt cfg.port or s.enabled isnt cfg.enabled
  
  saveProto: (proto, cb) ->
    obj = @getProto proto
    theCb = =>
      @config[proto] = obj
      cb()
    if proto is 'http'
      window.goule.api.setHttp obj, theCb
    else
      window.goule.api.setHttps obj, theCb
  
  getTimeout: ->
    num = parseInt @timeoutInput.val()
    return 0 if isNaN num
    return num
  
  didTimeoutChange: ->
    return @getTimeout() != @config.admin.session_timeout
  
  saveTimeout: (cb) ->
    to = @getTimeout()
    window.goule.api.setSessionTimeout to, =>
      @config.admin.session_timeout = to
      cb()
  
  getProxy: ->
    dict =
      rewrite_host: @rewriteInput.prop 'checked'
      websockets: @websocketsInput.prop 'checked'
    return dict
  
  didProxyChange: ->
    s = @getProxy()
    cfg = @config.proxy
    return s.websockets isnt cfg.websockets or
      s.rewrite_host isnt cfg.rewrite_host
  
  saveProxy: (cb) ->
    obj = @getProxy()
    window.goule.api.setProxy obj, =>
      @config.proxy = obj
      cb()

class PasswordChanger
  constructor: ->
    @all = $ '#as-chpass-form input'
    @passwordInput = $ '#as-new-password'
    @confirmInput = $ '#as-confirm-password'
    $('#as-chpass-form').submit (e) =>
      e.preventDefault()
      @submit()
  
  disable: -> @all.css opacity: '0.5', 'pointer-events': 'none'
  
  enable: -> @all.css opacity: '1.0', 'pointer-events': 'auto'
  
  submit: ->
    newPass = @passwordInput.val()
    confirm = @confirmInput.val()
    if newPass isnt confirm
      @confirmInput.effect 'shake'
      return
    @disable()
    window.goule.api.changePassword newPass, =>
      @enable()
      @passwordInput.val ''
      @confirmInput.val ''

class ForwardRules
  constructor: ->
    @showingSave = false
    @configRules = []
    @listDiv = $ '#as-rule-list'
    @container = $ '#as-rules'
    @addButton = $ '#as-rules .inner-add-button'
    @saveButton = $ '#as-rules .save-button'
    @addButton.click =>
      el = @addRule hostname: '', scheme: 'http', path: ''
      el.find('input').focus()
      @inputChanged()
    @saveButton.click => @save()
  
  disable: -> @container.css opacity: '0.5', 'pointer-events': 'none'
  
  enable: -> @container.css opacity: '1.0', 'pointer-events': 'auto'

  inputChanged: ->
    if @didChange()
      return if @showingSave
      @showingSave = true
      @saveButton.stop true, true
      @addButton.stop true, true
      @saveButton.css left: '380px', opacity: '0.0', display: 'inline-block'
      @saveButton.animate left: '405px', opacity: '1.0'
      @addButton.animate left: '355px'
    else
      return if not @showingSave
      @showingSave = false
      @saveButton.stop true, true
      @addButton.stop true, true
      @saveButton.animate {left: '380px', opacity: '0.0'}, =>
        @saveButton.css display: 'none'
      @addButton.animate left: '380px'

  update: (config) ->
    @configRules = config.admin.rules
    @listDiv.empty()
    @addRule rule for rule in @configRules
    @inputChanged()
  
  save: ->
    @disable()
    rules = @getRules()
    window.goule.api.setAdminRules rules, =>
      @enable()
      @configRules = rules
      @inputChanged()
  
  didChange: ->
    rules = @getRules()
    if rules.length isnt @configRules.length
      return true
    for rule, i in rules
      if not ForwardRules.rulesEqual rule, @configRules[i]
        return true
    return false
  
  getRules: ->
    res = []
    for input in @listDiv.find 'input'
      # parse the URL to form a rule
      res.push ForwardRules.parseRule $(input).val()
    return res
  
  addRule: (rule) ->
    el = $ '<div class="as-rule"><input /><button></button></div>'
    input = el.find 'input'
    input.val rule.scheme + "://" + rule.hostname + rule.path
    input.on 'input', => @inputChanged()
    el.find('button').click =>
      el.remove()
      @inputChanged()
    @listDiv.append el
    return el
  
  @parseRule: (ruleStr) ->
    # A rule like "http://localhost/foo"
    match = /^(http|https):\/\/(.*?)\/(.*)$/.exec ruleStr
    if match?
      return scheme: match[1], hostname: match[2], path: '/' + match[3]
    # A rule like "http://google.com"
    match = /^(http|https):\/\/(.*)$/.exec ruleStr
    if match?
      return scheme: match[1], hostname: match[2], path: ''
    # A rule like "aqnichol.com/foo"
    match = /^(.*)\/(.*)$/.exec ruleStr
    if match?
      return scheme: 'http', hostname: match[1], path: '/' + match[2]
    # A rule like "aqnichol.com"
    return scheme: 'http', hostname: ruleStr, path: ''
  
  @rulesEqual: (r1, r2) ->
    for own key, val of r1
      return false if r2[key] isnt val
    return true

$ -> window.goule.adminSettings = new AdminSettings()
