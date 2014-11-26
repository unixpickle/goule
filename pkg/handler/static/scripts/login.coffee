window.goule = {} if not window.goule?

window.goule.showLogin = (animate = false) ->
	if animate
		$('#login').css display: 'inline-block', opacity: '0'
		$('#controls').fadeOut()
		$('#login').fadeIn()
	else
		$('#login').css display: 'inline-block'
	$('#login-input').focus()

$ ->
	window.goule.showLogin()
	$('#login-input').keyup ->
		size = 90 - 5 * $('#login-input').val().length
		size = 30 if size < 30
		$('#login-input').css 'font-size': "#{size}px"
