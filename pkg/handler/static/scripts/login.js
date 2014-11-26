// Generated by CoffeeScript 1.7.1
(function() {
  if (window.goule == null) {
    window.goule = {};
  }

  window.goule.showLogin = function(animate) {
    if (animate == null) {
      animate = false;
    }
    if (animate) {
      $('#login').css({
        display: 'inline-block',
        opacity: '0'
      });
      $('#controls').fadeOut();
      $('#login').fadeIn();
    } else {
      $('#login').css({
        display: 'inline-block'
      });
    }
    return $('#login-input').focus();
  };

  $(function() {
    window.goule.showLogin();
    return $('#login-input').keyup(function() {
      var size;
      size = 90 - 5 * $('#login-input').val().length;
      if (size < 30) {
        size = 30;
      }
      return $('#login-input').css({
        'font-size': "" + size + "px"
      });
    });
  });

}).call(this);
