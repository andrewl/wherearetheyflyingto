/* eslint-disable  func-names */
/* eslint quote-props: ["error", "consistent"]*/

/*
 * This is a lambda to be used for an Alexa voice skill to return the
 * details of the last plane that flew over
 */


'use strict';

const Alexa = require('alexa-sdk');
const doc = require('dynamodb-doc');
const dynamo = new doc.DynamoDB();

const APP_ID = 'where_are_they_flying_to';

const languageStrings = {
    'en': {
        translation: {
            SKILL_NAME: 'Where Are They Flying To',
            DESTINATION: "The last flight over was going to ",
            ALTITUDE: " at an altitude of ",
            FEET: " feet",
            AT: " at ",
            HELP_MESSAGE: 'I can tell you details of the last plane that flew over',
            STOP_MESSAGE: 'Goodbye!',
        },
    }
};

var flight_details = {};

var getFlightDetails = function(alexa) {
        dynamo.query({
            ExpressionAttributeValues: {
   ":v1": "A flight just passed overhead",
   ":v2": "201",
  }, 
            TableName: 'watft',
            KeyConditionExpression: 'msg = :v1 AND ts > :v2',
            Limit: 1,
            ScanIndexForward: false
        },
        function(err, res) {
            if (err) {
                flight_details.err = err;
            }
            else {
                flight_details = res.Items[0];
                flight_details.err = '';
            }
            alexa.execute();
        })

};

const handlers = {
    'LaunchRequest': function () {
        this.emit('GetFact');
    },
    'GetNewFactIntent': function () {
        this.emit('GetFact');
    },
    'GetFact': function () {
        var message;
          if (flight_details.err == '') {
            var flight_time = new Date(flight_details.time);  
            message = this.t('DESTINATION') + flight_details.destination_name + this.t('ALTITUDE') + flight_details.altitude + this.t('FEET') + this.t('AT') + flight_time.getHours() + " " + flight_time.getMinutes();
          }
          else {
              message = "There was an error " + JSON.stringify(flight_details.err);
          }
          const speechOutput = message;
          this.emit(':tellWithCard', speechOutput, this.t('SKILL_NAME'), message);
    },
    'AMAZON.HelpIntent': function () {
        const speechOutput = this.t('HELP_MESSAGE');
        const reprompt = this.t('HELP_MESSAGE');
        this.emit(':ask', speechOutput, reprompt);
    },
    'AMAZON.CancelIntent': function () {
        this.emit(':tell', this.t('STOP_MESSAGE'));
    },
    'AMAZON.StopIntent': function () {
        this.emit(':tell', this.t('STOP_MESSAGE'));
    },
};

exports.handler = function (event, context) {
    const alexa = Alexa.handler(event, context);
    alexa.APP_ID = APP_ID;
    alexa.resources = languageStrings;
    alexa.registerHandlers(handlers);
    getFlightDetails(alexa);
};

