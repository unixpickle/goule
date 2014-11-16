class Settings
	constructor: ->
	
	show: ->

$ ->
	window.goule = {} if not window.goule?
	window.goule.settings = new Settings()
