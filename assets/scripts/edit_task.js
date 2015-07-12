(function() {

  $(function() {
    var editor = new window.TaskEditor($('#edit-task-fields'), window.taskData);
    $('#save').click(function() {
      // Honestly, this is because I'm too lazy to use AJAX.
      var taskJSON = JSON.stringify(editor.getTask());
      var $form = $('<form method="POST"><input name="task" type="hidden"></form>');
      $form.find('input').val(taskJSON);
      $form.submit();
    });
    $('#cancel').click(function() {
      window.location = '/';
    });
  });

})();
