require "logstash/filters/grok"

describe LogStash::Filters::Grok do
  extend LogStash::RSpec
  config <<-CONFIG

    #{File.read("patterns/firehose.conf")}

  CONFIG

  describe "Parse doppler message" do
    sample('message' => '{"cf_app_id":"918676fd-0628-422a-a6b2-3aa467e2f0ba","level":"info","message_type":"ERR","msg":"172.18.11.137, 10.230.16.68 - - [16/Dec/2014 14:10:54] \"GET /assets/application.5d41646dbe3a6a844dd91f864db21f5c.css HTTP/1.1\" 304 - 0.0005","source_instance":"0","source_type":"App","time":"2014-12-16T15:10:54+01:00"}',
           'program' => 'doppler') do
      insist { subject['cf_app_id'] } == '918676fd-0628-422a-a6b2-3aa467e2f0ba'
      insist { subject['message_type'] } == 'ERR'
      insist { subject['source_type'] } == 'App'
      insist { subject['source_instance'] } == '0'
      insist { subject['time'] } == '2014-12-16T15:10:54+01:00'
      insist { subject['msg'] } == '172.18.11.137, 10.230.16.68 - - [16/Dec/2014 14:10:54] "GET /assets/application.5d41646dbe3a6a844dd91f864db21f5c.css HTTP/1.1" 304 - 0.0005'
    end
  end

  describe "Parse doppler message" do
    sample('message' => '{"cf_app_id":"8a3849fa-3f4c-44ef-bf36-92095417f204","level":"info","message_type":"ERR","msg":"\u0009at com.googlecode.utterlyidle.httpserver.RestHandler.handle(RestHandler.java:33)","source_instance":"0","source_type":"App","time":"2014-12-16T15:20:10+01:00"}',
           'program' => 'doppler') do
      insist { subject['cf_app_id'] } == '8a3849fa-3f4c-44ef-bf36-92095417f204'
      insist { subject['message_type'] } == 'ERR'
      insist { subject['source_type'] } == 'App'
      insist { subject['source_instance'] } == '0'
      insist { subject['msg'] } == "\tat com.googlecode.utterlyidle.httpserver.RestHandler.handle(RestHandler.java:33)"
      insist { subject['time'] } == '2014-12-16T15:20:10+01:00'
    end
  end
end
