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

    this._initializeArguments(task);
    this._initializeEnvironment(task);
    this._initializeFields(task);
  }

  TaskEditor.prototype._addArgument = function() {
    this._$arguments.append(createArgumentElement(''));
  };

  TaskEditor.prototype._addEnv = function() {
    this._$env.append(createEnvironmentElement('', ''));
  };

  TaskEditor.prototype._getField = function(name) {
    return this._$fields.find('.task-editor-' + name);
  };

  TaskEditor.prototype._initializeArguments = function(task) {
    var $argumentsTitle = $('<div class="task-editor-heading">' +
      '<h1 class="task-editor-title">Arguments</h1>' +
      '<button class="task-editor-add-argument-button task-editor-add-button">Add</button></div>');
    $argumentsTitle.find('.task-editor-add-argument-button').click(this._addArgument.bind(this));

    this._$arguments = $('<div></div>');
    for (var i = 0, len = task.Args.length; i < len; ++i) {
      this._$arguments.append(createArgumentElement(task.Args[i]));
    }

    this._$container.append($argumentsTitle, this._$arguments);
  };

  TaskEditor.prototype._initializeEnvironment = function(task) {
    var $envTitle = $('<div class="task-editor-heading">' +
      '<h1 class="task-editor-title">Environment</h1>' +
      '<button class="task-editor-add-env-button task-editor-add-button">Add</button></div>');
    $envTitle.find('.task-editor-add-env-button').click(this._addEnv.bind(this));

    this._$env = $('<div></div>');
    var keys = Object.keys(task.Env).sort();
    for (var i = 0; i < keys.length; ++i) {
      this._$env.append(createEnvironmentElement(keys[i], task.Env[i]));
    }

    this._$container.append($envTitle, this._$env);
  };

  TaskEditor.prototype._initializeFields = function(task) {
    this._$fields = $('<div class="task-editor-fields">' +

      '<label class="field-label">Directory</label>' +
      '<input class="task-editor-directory">' +

      '<label class="field-label">Auto-launch</label>' +
      '<input class="task-editor-auto-launch" type="checkbox"><br>' +

      '<label class="field-label">Auto-relaunch</label>' +
      '<input class="task-editor-auto-relaunch" type="checkbox"><br>' +

      '<div class="task-editor-relaunch-interval-field">' +
      '<label class="field-label task-editor">Relaunch interval (sec)</label>' +
      '<input class="task-editor-relaunch-interval"></div>' +

      '<label class="field-label">Set GID</label>' +
      '<input class="task-editor-set-gid" type="checkbox"><br>' +

      '<div class="task-editor-gid-field">' +
      '<label class="field-label">GID</label>' +
      '<input class="task-editor-gid"></div>' +

      '<label class="field-label">Set UID</label>' +
      '<input class="task-editor-set-uid" type="checkbox"><br>' +

      '<div class="task-editor-uid-field">' +
      '<label class="field-label">UID</label>' +
      '<input class="task-editor-uid"></div>' +

      '</div>');

    this._registerFieldEvents();
    this._updateFieldsFromTask(task);
    this._$container.append(this._$fields);
  };

  TaskEditor.prototype._registerFieldEvents = function() {
    var checkFields = ['auto-launch', 'auto-relaunch', 'set-gid', 'set-uid'];
    for (var i = 0; i < checkFields.length; ++i) {
      this._getField(checkFields[i]).change(this._updateFieldVisibility.bind(this));
    }
  };

  TaskEditor.prototype._updateFieldVisibility = function() {
    var fields = {
      'auto-relaunch': 'relaunch-interval',
      'set-gid': 'gid',
      'set-uid': 'uid'
    };
    var keys = Object.keys(fields);
    for (var i = 0; i < keys.length; ++i) {
      var checkField = keys[i];
      var display = 'none';
      if (this._getField(checkField).is(':checked')) {
        display = 'block';
      }
      this._getField(fields[checkField] + '-field').css({display: display});
    }
  };

  TaskEditor.prototype._updateFieldsFromTask = function(task) {
    this._getField('auto-launch').attr('checked', task.AutoRun);
    this._getField('auto-relaunch').attr('checked', task.Relaunch);
    this._getField('set-gid').attr('checked', task.SetGID);
    this._getField('set-uid').attr('checked', task.SetUID);
    this._getField('directory').val(task.Dir);
    this._getField('relaunch-interval').val(task.Interval);
    this._getField('gid').val(task.GID);
    this._getField('uid').val(task.UID);
    this._updateFieldVisibility();
  };

  function createArgumentElement(arg) {
    var $res = $('<div class="task-editor-argument"><input placeholder="Argument">' +
      '<button>Remove</button></div>');
    $res.find('input').val(arg);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  function createEnvironmentElement(name, value) {
    var $res = $('<div><input class="task-editor-env-key"><input class="task-editor-env-value">' +
      '<button>Remove</button></div>');
    $res.find('task-editor-env-key').val(name);
    $res.find('task-editor-env-value').val(value);
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
