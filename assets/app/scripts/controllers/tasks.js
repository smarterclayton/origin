'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:TasksController
 * @description
 * # TasksController displays tasks from TaskService
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('TasksController', function ($scope, TaskList) {
    $scope.tasks = function() {
      return TaskList.taskList();
    };
    $scope.delete = function(task) {
      task.toDelete = true;
    };
    $scope.expanded = false;
    $scope.toggleExpand = function() {
      $scope.expanded = !$scope.expanded;
    };
  });
