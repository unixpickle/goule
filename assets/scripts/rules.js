(function() {

  function createRuleElement(host, targets) {
  }

  function createTargetElement(hostname) {
    var element = $('<div />', {class: 'target'});
    var input = $('<input />', {value: hostname});
    var remove = $('<button>Remove</button>');
    element.append(input);
    element.append(remove);
    remove.click(function() {
      element.remove();
    });
    return element;
  }

  function loadRules(rules) {
    for (var key in rules) {
      if (!rules.hasOwnProperty(key)) {
        continue;
      }
      var targets = rules[key];
      var element = createRuleElement(key, targets);
    }
  }

  

  window.app.loadRules = loadRules;

})();
