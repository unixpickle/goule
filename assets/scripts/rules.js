(function() {

  function addRule() {
    var container = $('#rules');
    container.append(createRuleElement('', ['']));
  }

  function createRuleElement(host, targets) {
    var element = $('<div />', {class: 'rule'});
    var input = $('<input />', {
      value: host,
      placeholder: 'Host'
    });
    var remove = $('<button>Remove</button>');
    var add = $('<button>Add Target</button>');
    element.append(input);
    element.append(remove);
    element.append(add);
    remove.click(function() {
      element.remove();
    });
    add.click(function() {
      element.append(createTargetElement(''));
    });
    for (var i = 0, len = targets.length; i < len; ++i) {
      element.append(createTargetElement(targets[i]));
    }
    return element;
  }

  function createTargetElement(hostname) {
    var element = $('<div />', {class: 'target'});
    var input = $('<input />', {
      value: hostname,
      placeholder: 'Target'
    });
    var remove = $('<button>Remove</button>');
    element.append(input);
    element.append(remove);
    remove.click(function() {
      element.remove();
    });
    return element;
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
    var container = $('#rules');
    for (var i = 0, len = hosts.length; i < len; ++i) {
      var key = hosts[i];
      var targets = rules[key];
      var element = createRuleElement(key, targets);
      container.push(element);
    }
  }

  window.app.addRule = addRule;
  window.app.loadRules = loadRules;

})();
