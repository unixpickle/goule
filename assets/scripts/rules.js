(function() {

  function addRule() {
    $('#rules').append(createRuleElement('', ['']));
  }

  function createRuleElement(host, targets) {
    var $element = $('<div></div>', {class: 'rule'});
    var $input = $('<input></input>', {
      value: host,
      placeholder: 'Host',
      class: 'host-name'
    });
    var $remove = $('<button>Remove</button>').click(function() {
      $element.remove();
    });
    var $add = $('<button>Add Target</button>').click(function() {
      $element.append(createTargetElement(''));
    });
    $element.append($input, $remove, $add);
    for (var i = 0, len = targets.length; i < len; ++i) {
      $element.append(createTargetElement(targets[i]));
    }
    return $element;
  }

  function createTargetElement(hostname) {
    var $element = $('<div></div>', {class: 'target'});
    var $input = $('<input></input>', {
      value: hostname,
      placeholder: 'Target',
      class: 'target-name'
    });
    var $remove = $('<button>Remove</button>').click(function() {
      $element.remove();
    });
    $element.append($input, $remove);
    return $element;
  }

  function loadRules(rules) {
    // Get a sorted list of hosts.
    var hosts = [];
    for (var key in rules) {
      if (!rules.hasOwnProperty(key)) {
        continue;
      }
      hosts.push(key);
    }
    hosts.sort();
    // Add an element for each host.
    var $container = $('#rules');
    for (var i = 0, len = hosts.length; i < len; ++i) {
      var key = hosts[i];
      var targets = rules[key];
      var $element = createRuleElement(key, targets);
      $container.append($element);
    }
  }

  function save() {
    // Generate a rules object.
    var result = {};
    var $rules = $('.rule');
    for (var i = 0, len = $rules.length; i < len; ++i) {
      var $rule = $rules.eq(i);
      var name = $rule.find('input.host-name').val();
      var targets = $rule.find('input.target-name');
      var targetNames = [];
      for (var j = 0, len1 = targets.length; j < len1; ++j) {
        targetNames[j] = targets.eq(j).val();
      }
      result[name] = targetNames;
    }

    // Send the rules object to the server.
    var encoded = encodeURIComponent(JSON.stringify(result));
    window.location = '/setrules?rules=' + encoded;
  }

  window.app.addRule = addRule;
  window.app.loadRules = loadRules;
  window.app.save = save;

})();
