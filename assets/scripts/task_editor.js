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
    task = (task || DEFAULT_TASK);
    this._$container = $container;

    this._initializeArguments(task);
    this._initializeEnvironment(task);
    this._initializeFields(task);
  }

  TaskEditor.prototype.getTask = function() {
    var env = {};
    var $values = this._$container.find('.task-editor-env-value');
    this._$container.find('.task-editor-env-key').each(function(i, element) {
      env[$(element).val()] = $values.eq(i).val();
    });

    var args = [];
    this._$container.find('.task-editor-argument input').each(function(i, element) {
      args.push($(element).val());
    });

    return {
      Args: args,
      AutoRun: this._getField('auto-launch').is(':checked'),
      Dir: this._getField('directory').val(),
      Env: env,
      GID: parseInt(this._getField('gid').val()) || 0,
      UID: parseInt(this._getField('uid').val()) || 0,
      SetGID: this._getField('set-gid').is(':checked'),
      SetUID: this._getField('set-uid').is(':checked'),
      Relaunch: this._getField('auto-relaunch').is(':checked'),
      Interval: parseInt(this._getField('relaunch-interval').val())
    };
  };

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
    var $argumentsTitle = $('<div class="field-set-action-heading">' +
      '<h1>Arguments</h1><button class="field-set-add-button">Add</button></div>');
    $argumentsTitle.find('.field-set-add-button').click(this._addArgument.bind(this));

    this._$arguments = $('<div></div>');
    for (var i = 0, len = task.Args.length; i < len; ++i) {
      this._$arguments.append(createArgumentElement(task.Args[i]));
    }

    this._$container.append($argumentsTitle, this._$arguments);
  };

  TaskEditor.prototype._initializeEnvironment = function(task) {
    var $envTitle = $('<div class="field-set-action-heading">' +
      '<h1>Environment</h1><button class="field-set-add-button">Add</button></div>');
    $envTitle.find('.field-set-add-button').click(this._addEnv.bind(this));

    this._$env = $('<div></div>');
    var keys = Object.keys(task.Env).sort();
    for (var i = 0; i < keys.length; ++i) {
      this._$env.append(createEnvironmentElement(keys[i], task.Env[i]));
    }

    this._$container.append($envTitle, this._$env);
  };

  TaskEditor.prototype._initializeFields = function(task) {
    this._$fields = $('<div class="task-editor-fields">' +

      '<div class="field">' +
      '<label class="input-field-label">Directory</label>' +
      '<input class="input-field-input task-editor-directory"></div>' +

      '<div class="field">' +
      '<label class="generic-field-label">Auto-launch</label>' +
      '<input class="generic-field-content task-editor-auto-launch" type="checkbox"></div>' +

      '<div class="field">' +
      '<label class="generic-field-label">Auto-relaunch</label>' +
      '<input class="generic-field-content task-editor-auto-relaunch" type="checkbox"></div>' +

      '<div class="field task-editor-relaunch-interval-field">' +
      '<label class="input-field-label task-editor">Relaunch interval (sec)</label>' +
      '<input class="input-field-input task-editor-relaunch-interval"></div>' +

      '<div class="field">' +
      '<label class="generic-field-label">Set GID</label>' +
      '<input class="generic-field-content task-editor-set-gid" type="checkbox"></div>' +

      '<div class="field task-editor-gid-field">' +
      '<label class="input-field-label">GID</label>' +
      '<input class="input-field-input task-editor-gid"></div>' +

      '<div class="field">' +
      '<label class="generic-field-label">Set UID</label>' +
      '<input class="generic-field-content task-editor-set-uid" type="checkbox"></div>' +

      '<div class="field task-editor-uid-field">' +
      '<label class="input-field-label">UID</label>' +
      '<input class="input-field-input task-editor-uid"></div>' +

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

  window.TaskEditor = TaskEditor;

})();
