(function() {

  $(function() {
    var editor = new window.TaskEditor($('#add-task-fields'));
    $('#submit').click(function() {
      var taskJSON = JSON.stringify(editor.getTask());
      postData('task', taskJSON);
    });
  });

})();
