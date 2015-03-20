'use strict';

angular.module('openshiftConsole')
.factory('LabelFilter', [function() {
  function LabelFilter() {
    this._existingLabels = {};
    this._labelSelector = new LabelSelector(null, true);
    this._onActiveFiltersChangedCallbacks = $.Callbacks();
  }

  LabelFilter.prototype.addLabelSuggestionsFromResources = function(items, map) {
    // check if we are extracting from a single item or a hash of items
    if (items.metadata && items.metadata.name) {
      this._extractLabelsFromItem(items, map);
    }
    else {
      var self = this;
      angular.forEach(items, function(item) {
        self._extractLabelsFromItem(item, map);
      });
    }
  };

  LabelFilter.prototype.setLabelSuggestions = function(suggestions) {
    this._existingLabels = suggestions;
  };

  LabelFilter.prototype._extractLabelsFromItem = function(item, map) {
    var labels = item.metadata ? item.metadata.labels : {};
    var self = this;
    angular.forEach(labels, function(value, key) {
      if (!map[key]) {
        map[key] = [];
      }
      map[key].push({value: value});
    });
  };

  LabelFilter.prototype.getLabelSelector = function() {
    return this._labelSelector;
  };

  LabelFilter.prototype.onActiveFiltersChanged = function(callback) {
    this._onActiveFiltersChangedCallbacks.add(callback);
  };

  // Creates the filtering widget input inside of filterInputElement
  // Creates the filtering widget active filters boxes inside of activeFiltersElement
  // filterInputElement and activeFiltersElement should be empty HTML nodes
  LabelFilter.prototype.setupFilterWidget = function(filterInputElement, activeFiltersElement) {
    var self = this;

    this._labelFilterRootElement = filterInputElement;
    this._labelFilterActiveFiltersRootElement = activeFiltersElement;

    // Render base select boxes and buttons for inputs of widget
    var labelFilterElem = $('<div>')
      .addClass("label-filter")
      .appendTo(filterInputElement);

    this._labelFilterKeyInput = $('<select>')
      .addClass("label-filter-key")
      .attr("placeholder", "Label key ")
      .appendTo(labelFilterElem);

    this._labelFilterOperatorInput = $('<select>')
      .addClass("label-filter-operator")
      .attr("placeholder", "matching(...)")
      .hide()
      .appendTo(labelFilterElem);      

    this._labelFilterValuesInput = $('<select>')
      .addClass("label-filter-values")
      .attr("placeholder", "Value(s)")
      .attr("multiple", true)
      .hide()
      .appendTo(labelFilterElem);   

    this._labelFilterAddBtn = $('<button>')
      .addClass("label-filter-add btn btn-default btn-lg disabled")
      .attr("disabled", true)
      .appendTo(filterInputElement)
      .append(
        $('<i>')
          .addClass("fa fa-plus")
      )
      .append(
        $('<span>')
          .text(" Filter")
      );

    // Render active filters area
    this._labelFilterActiveElement = $('<span>')
      .addClass("label-filter-active")
      .hide()
      .appendTo(activeFiltersElement)
      .append(
        $('<a>')
          .addClass("label-filtering-remove-all label label-primary")
          .prop("href", "javascript:;")
          .append(
            $('<i>')
              .addClass("fa fa-filter")
              .css("padding-right", "5px")
          )
          .append(
            $('<span>')
              .text("Clear all filters")
          )
          .append(
            $('<i>')
              .addClass("fa fa-times")
          )
      ).click(function() {
        $(this).hide();
        self._labelFilterActiveFiltersElement.empty();
        self._clearActiveFilters();
      });

    this._labelFilterActiveFiltersElement = $('<span>')
      .addClass("label-filter-active-filters")
      .appendTo(activeFiltersElement);

    // Create selectize widgets for the select fields and wire them together
    this._labelFilterKeyInput.selectize({
      valueField: "key",
      labelField: "key",
      searchField: ["key"],
      create: true,
      persist: true, // i want this to be false but there appears to be a bug in selectize where setting
                     // this to false has a side effect of causing items that were not created by the user
                     // to also disappear from the list after being removed
      preload: true,
      onItemAdd: function(value, $item) {
        var selectizeValues = self._labelFilterValuesSelectize;
        selectizeValues.clearOptions();
        selectizeValues.load(function(callback) {
          var options = [];
          var key = self._labelFilterKeySelectize.getValue();
          if (!key) {
            return options;
          }
          var optionsMap = self._existingLabels;
          // if there are no values for this key, like when user chooses to explicitly add a key
          // then there are no values to suggest
          if (!optionsMap[key]) {
            callback({});
            return;
          }
          //for each value for key
          for (var i = 0; i < optionsMap[key].length; i++) {                  
            options.push(optionsMap[key][i]);
          }                
          callback(options);
        });          

        self._labelFilterOperatorSelectizeInput.css("display", "inline-block");
        var operator = self._labelFilterOperatorSelectize.getValue();
        if (!operator) {
          self._labelFilterOperatorSelectize.focus();
        }                
        else {
          selectizeValues.focus();
        }
      },
      onItemRemove: function(value) {
        self._labelFilterOperatorSelectizeInput.hide();
        self._labelFilterOperatorSelectize.clear();
        self._labelFilterValuesSelectizeInput.hide();
        self._labelFilterValuesSelectize.clear();
        self._labelFilterAddBtn.addClass("disabled").prop('disabled', true);
      },
      load: function(query, callback) {
        var options = [
        ];
        var keys = Object.keys(self._existingLabels);
        for (var i = 0; i < keys.length; i++) {
          options.push({
            key: keys[i]
          });
        }                
        callback(options);
      }
    });

    this._labelFilterKeySelectize = this._labelFilterKeyInput.prop("selectize");
    this._labelFilterKeySelectizeInput = $('.selectize-control.label-filter-key', labelFilterElem);

    this._labelFilterOperatorInput.selectize({
      valueField: "type",
      labelField: "label",
      searchField: ["label"],
      options: [
        {type: "exists", label: "exists"},
        {type: "in", label: "in ..."},
        {type: "not in", label: "not in ..."}
      ],
      onItemAdd: function(value, $item) {
        // if we selected "exists" enable the add button and stop here
        if (value == "exists") {
          self._labelFilterAddBtn.removeClass("disabled").prop('disabled', false).focus();
          return;
        }

        // otherwise
        self._labelFilterValuesSelectizeInput.css("display", "inline-block");
        self._labelFilterValuesSelectize.focus();
      },
      onItemRemove: function(value) {
        self._labelFilterValuesSelectizeInput.hide();
        self._labelFilterValuesSelectize.clear();
        self._labelFilterAddBtn.addClass("disabled").prop('disabled', true);
      }
    });

    this._labelFilterOperatorSelectize = this._labelFilterOperatorInput.prop("selectize");
    this._labelFilterOperatorSelectizeInput = $('.selectize-control.label-filter-operator', labelFilterElem);
    this._labelFilterOperatorSelectizeInput.hide();

    this._labelFilterValuesInput.selectize({
      valueField: "value",
      labelField: "value",
      searchField: ["value"],
      plugins: ['remove_button'],
      create: true,
      persist: true, // i want this to be false but there appears to be a bug in selectize where setting
                     // this to false has a side effect of causing items that were not created by the user
                     // to also disappear from the list after being removed
      preload: true,
      onItemAdd: function(value, $item) {
        self._labelFilterAddBtn.removeClass("disabled").prop('disabled', false);
      },
      onItemRemove: function(value) {
        // disable button if we have removed all the values                
      },
      load: function(query, callback) {
        var options = [];
        var key = self._labelFilterKeySelectize.getValue();
        if (!key) {
          return options;
        }
        var optionsMap = self._existingLabels;
        // if there are no values for this key, like when user chooses to explicitly add a key
        // then there are no values to suggest
        if (!optionsMap[key]) {
          callback({});
          return;
        }        
        //for each value for key
        for (var i = 0; i < optionsMap[key].length; i++) {                  
          options.push(optionsMap[key][i]);
        }                
        callback(options);
      }
    });

    this._labelFilterValuesSelectize = this._labelFilterValuesInput.prop("selectize");
    this._labelFilterValuesSelectizeInput = $('.selectize-control.label-filter-values', labelFilterElem);
    this._labelFilterValuesSelectizeInput.hide();

    this._labelFilterAddBtn.click(function() {
      // grab the values before clearing out the fields
      var key = self._labelFilterKeySelectize.getValue();
      var operator = self._labelFilterOperatorSelectize.getValue();
      var values = self._labelFilterValuesSelectize.getValue();

      self._labelFilterKeySelectize.clear();
      self._labelFilterOperatorSelectizeInput.hide();
      self._labelFilterOperatorSelectize.clear();
      self._labelFilterValuesSelectizeInput.hide();
      self._labelFilterValuesSelectize.clear();
      self._labelFilterAddBtn.addClass("disabled").prop('disabled', true);              

      // show the filtering active indicator and add the individual filter to the list of active filters
      self._labelFilterActiveElement.show();
      self._addActiveFilter(key, operator, values);
    });

    // If we are transitioning scenes we may still have filters active but be re-creating the DOM for the widget
    if (!this._labelSelector.isEmpty()) {
      this._labelFilterActiveElement.show();
      this._labelSelector.each(function(filter) {
        self._renderActiveFilter(filter);
      });
    }      
  };



  LabelFilter.prototype._addActiveFilter = function(key, operator, values) {
    var filter = this._labelSelector.addConjunct(key, operator, values);
    this._onActiveFiltersChangedCallbacks.fire(this._labelSelector);  
    this._renderActiveFilter(filter);
  };

  LabelFilter.prototype._renderActiveFilter = function(filter) {
    // render the new filter indicator
    $('<a>')
      .addClass("label label-default label-filter-active-filter")
      .prop("href", "javascript:;")
      .prop("filter-label-id", filter.id)
      .click($.proxy(this, '_removeActiveFilter'))
      .append(
        $('<span>')
          .text(filter.string)
          // TODO move to the less styles instead
          .css("padding-right", "5px")
      )
      .append(
        $('<i>')
          .addClass("fa fa-times")
      )
      .appendTo(this._labelFilterActiveFiltersElement);  
  };

  LabelFilter.prototype._removeActiveFilter = function(e) {
    var filterElem = $(e.target).closest('.label-filter-active-filter');
    var filter = filterElem.prop("filter-label-id");
    filterElem.remove();
    if($('.label-filter-active-filter', this._labelFilterActiveFiltersElement).length == 0) {
      this._labelFilterActiveElement.hide();
    }

    this._labelSelector.removeConjunct(filter);
    this._onActiveFiltersChangedCallbacks.fire(this._labelSelector);
  };

  LabelFilter.prototype._clearActiveFilters = function() {
    this._labelSelector.clearConjuncts();
    this._onActiveFiltersChangedCallbacks.fire(this._labelSelector);
  };

  LabelFilter.prototype.toggleFilterWidget = function(show) {
    if (this._labelFilterRootElement) {
      if (show) {
        this._labelFilterRootElement.show();
      }
      else {
        this._labelFilterRootElement.hide();
      }
    }
    if (this._labelFilterActiveFiltersRootElement) {
      if (show) {
        this._labelFilterActiveFiltersRootElement.show();
      }
      else {
        this._labelFilterActiveFiltersRootElement.hide();
      }
    }
  };

  return new LabelFilter();
}]);