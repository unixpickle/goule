class Tabs
	constructor: ->
		@element = $ '#active-tab-indicator'
		@showing = ''
	
	change: (button, animate = true) ->
		@showing = button[0].id
		if animate
			attributes =
				'width': button.outerWidth(),
				'left': button.offset().left
			@element.animate attributes, 'fast', ->
		else
			@element.css
				'width': button.outerWidth() + 'px',
				'left': button.offset().left + 'px'
  
	_setupButton: (name, showFunc) ->
		tab = $('#' + name)
		tab.click =>
			return if @showing is name
			showFunc()
			@change tab

$ ->
	window.goule = {} if not window.goule?
	tabs = new Tabs()
	window.goule.tabs = tabs
	tabs.change $('#list-tab'), false
	tabs._setupButton 'list-tab', -> window.goule.list.show
	tabs._setupButton 'settings-tab', -> window.goule.settings.show
