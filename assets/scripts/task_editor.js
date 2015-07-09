(function() {

  var DEFAULT_TASK = {
    Args: [''],
    AutoRun: false,
    Dir: '/',
    Env: {},
    GID: 0,
    UID: 0,
    SetGID: false,
    SetUID: false,
    Relaunch: false,
    Interval: 60
  };

  function TaskEditor($container, task) {
    this._$container = $container;

    this._$argumentsTitle = $('<div class="arguments-heading">' +
      '<h1 class="arguments-title">Arguments</h1>' +
      '<button class="add-argument-button">Add</button></div>');
    this._$container.append(this._$argumentsTitle);
    this._$argumentsTitle.find('.add-argument-button').click(this._addArgument.bind(this));

    this._$arguments = $('<div></div>');
    this._$container.append(this._$arguments);
    for (var i = 0, len = task.Args.length; i < len; ++i) {
      this._$arguments.append(createArgumentElement(task.Args[i]));
    }
  }

  TaskEditor.prototype._addArgument = function() {
    this._$arguments.append(createArgumentElement(''));
  };

  function createArgumentElement(arg) {
    var $res = $('<div class="task-argument"><input placeholder="Argument">' +
      '<button>Remove</button></div>');
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  function createTaskEditor($container, optionalTask) {
    new TaskEditor($container, optionalTask || DEFAULT_TASK)
  }

  window.createTaskEditor = createTaskEditor;

})();
