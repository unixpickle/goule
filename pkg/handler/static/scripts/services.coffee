window.goule = {} if not window.goule?

class Services
  show: (animate) ->
    if animate
      $('#add-service-button').fadeIn()
    else
      $('#add-service-button').css display: 'block', opacity: '1.0'
  
  hide: (animate) ->
    if animate
      $('#add-service-button').fadeOut()
    else
      $('#add-service-button').css display: 'none'

window.goule.services = new Services()
