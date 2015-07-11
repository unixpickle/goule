(function() {

  $(function() {
    var editor = new window.TaskEditor($('#add-task-fields'));
    $('#submit').click(function() {
      // Honestly, this is because I'm too lazy to use AJAX.
      var taskJSON = JSON.stringify(editor.getTask());
      var $form = $('<form method="POST"><input name="task" type="hidden"></form>');
      $form.find('input').val(taskJSON);
      $form.submit();
    });
  });

})();
