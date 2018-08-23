(function() {

  $(function() {
    var editor = new window.TaskEditor($('#edit-task-fields'), window.taskData);
    $('#save').click(function() {
      // Honestly, this is because I'm too lazy to use AJAX.
      var taskJSON = JSON.stringify(editor.getTask());
      postData('task', taskJSON)
    });
    $('#cancel').click(function() {
      window.location = '/';
    });
  });

})();
