window.goule = {} if not window.goule?

window.goule.headerControls = {}

servicesSel = true
tabServices = null
tabSettings = null

window.goule.headerControls.show = (animate) ->
  if animate
    $('#header-controls').fadeIn()
  else
    $('#header-controls').css display: 'block', opacity: '1.0'
  servicesSel = true
  selectTab tabServices, false

window.goule.headerControls.hide = (animate) ->
  if animate
    $('#header-controls').fadeOut()
  else
    $('#header-controls').css display: 'none'

selectTab = (tab, animate) ->
  width = tab.outerWidth()
  right = $(window).outerWidth() - (width + tab.offset().left) - 10
  if animate
    attributes = 'right': right, 'width': width
    $('#tab-selector').animate attributes, 'fast', ->
  else
    $('#tab-selector').css 'width': width + 'px', 'right': right + 'px'

$ ->
  $('.generate').mouseenter (e) ->
    number = Math.floor (Math.random() * (8192 - 1024)) + 1024
    $('.generate .random').html '' + number
  tabServices = $ '#tab-services'
  tabSettings = $ '#tab-settings'
  tabServices.click ->
    return if servicesSel
    servicesSel = true
    selectTab tabServices, true
    return false
  tabSettings.click ->
    return if not servicesSel
    servicesSel = false
    selectTab tabSettings, true
    return false
