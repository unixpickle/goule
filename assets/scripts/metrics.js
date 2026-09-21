(function() {

  $(function() {
    const editor = $('#metrics-viewer');
    editor.text(JSON.stringify(window.metricsData));
  });

})();
