(function() {

  $(function() {
    var $content = $('.main-content');
    if (window.backlog.length === 0) {
      $content.append('<label class="no-messages">No log messages</label>');
      return;
    }
    $content.css({visibility: 'hidden'});
    for (var i = 0, len = window.backlog.length; i < len; ++i) {
      var entry = window.backlog[i];
      var $row = $('<div class="entry"><label class="date"></label>' +
        '<label class="message"></label></div>');
      $row.find('.date').text(formatTimestamp(entry.Time));
      $row.find('.message').text(entry.Data);
      $row.addClass(['stdout', 'stderr', 'status'][entry.Type]);
      $content.append($row);
    }
    setTimeout(function() {
      $(document).scrollTop($(document).height());
      $content.css({visibility: 'visible'});
    }, 100);
  });

  function formatTime(millis) {
    var date = new Date(millis);
    var h = date.getHours();
    var m = date.getMinutes();
    if (m < 10) {
      m = '0' + m;
    }
    var s = date.getSeconds();
    if (s < 10) {
      s = '0' + s;
    }
    return h + ':' + m + ':' + s;
  }

  function formatTimestamp(millis) {
    var date = new Date(millis);
    return (date.getMonth()+1) + "/" + date.getDate() + "/" + date.getFullYear() + " " +
      formatTime(millis);
  }

})();
