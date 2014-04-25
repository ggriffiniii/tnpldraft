'use strict';

describe('Directive: draftHistory', function () {

  // load the directive's module
  beforeEach(module('tnpldraftApp'));
	beforeEach(module('scripts/directives/drafthistory.html'));

  var element,
    scope,
		draft;

  beforeEach(inject(function ($compile, $rootScope, draftState) {
    scope = $rootScope.$new();
		element = angular.element("<draft-history></draft-history>");
		$compile(element)(scope);
		scope.$digest();
		draft = draftState;
  }));

	it('should start with just header row', function() {
		var table = element.find('table');
		expect(table.length).toBe(1);
		expect(table.find('tr').length).toBe(1);
	});

	it('should have a row for every pick', function() {
		var table = element.find('table');
		expect(table.length).toBe(1);
		draft.picks.push({
			player: {
				id: 1,
				firstname: "Mike",
				lastname: "Trout",
				mlbteam: "Angels",
				positions: ["OF","U"],
				salary: 50
			},
			offering_team: 1,
			winning_team: 2
		});
		scope.$digest();
		expect(table.find('tr').length).toBe(2);

		draft.picks.push({
			player: {
				id: 2,
				firstname: "Miguel",
				lastname: "Cabrera",
				mlbteam: "Tigers",
				positions: ["3B","U"],
				salary: 150
			},
			offering_team: 2,
			winning_team: 1
		});
		scope.$digest();
		expect(table.find('tr').length).toBe(3);
	});
});
