(function() {

  $(function() {
    var editor = new window.TaskEditor($('#add-task-fields'));
    $('#submit').click(function() {
      var taskJSON = JSON.stringify(editor.getTask());
      var $form = $('<form method="POST"><input name="task" type="hidden"></form>');
      $form.find('input').val(taskJSON);
      $form.submit();
    });
  });

})();
