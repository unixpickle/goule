window.goule = {} if not window.goule?

window.goule.headerControls = {}

window.goule.headerControls.show = (animate) ->
  if animate
    $('#header-controls').fadeIn()
  else
    $('#header-controls').css display: 'block', opacity: '1.0'

window.goule.headerControls.hide = (animate) ->
  if animate
    $('#header-controls').fadeOut()
  else
    $('#header-controls').css display: 'none'