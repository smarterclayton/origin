'use strict';

angular.module('openshiftConsole')
  // This triggers when an element has either a toggle or data-toggle attribute set on it
  .directive('templateOptions', function() {
    return {
      restrict: 'E',
      templateUrl: 'views/_templateopt.html'
    };
  });
