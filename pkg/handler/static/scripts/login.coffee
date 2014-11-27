window.goule = {} if not window.goule?

window.goule.showLogin = ->
	$('#login').css display: 'inline-block', opacity: '0'
	$('#controls').fadeOut()
	$('#login').fadeIn()

$ ->
	$('#login-input').focus()
	$('#login-form').submit (e) ->
		e.preventDefault()
		$('#login-input').prop 'disabled', true
		window.goule.auth $('#login-input').val(), (succ) ->
			$('#login-input').prop 'disabled', false
			if not succ
				$('#login-input').effect 'shake'