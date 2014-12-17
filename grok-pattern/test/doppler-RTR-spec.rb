require "logstash/filters/grok"

describe LogStash::Filters::Grok do
  extend LogStash::RSpec
  config <<-CONFIG

    #{File.read("patterns/firehose.conf")}

  CONFIG

  describe "Parse doppler message" do
    sample('message' => '{"cf_app_id":"f5409d35-755b-4af8-865e-e6da42b43696","level":"info","message_type":"OUT","msg":"product-page-qa-online.dev.cf.i.domain.com - [16/12/2014:14:31:55 +0000] \"GET /internal/status HTTP/1.1\" 200 6 \"-\" \"-\" 10.230.16.68:36015 x_forwarded_for:\"192.87.127.242, 192.87.127.242, 10.9.0.188, 10.230.16.68\" vcap_request_id:5e565011-6935-4afa-5c9f-7597898eec55 response_time:0.005822253 app_id:f5409d35-755b-4af8-865e-e6da42b43696\n","source_instance":"0","source_type":"RTR","time":"2014-12-16T15:31:55+01:00"}',
           'program' => 'doppler') do
      insist { subject['cf_app_id'] } == 'f5409d35-755b-4af8-865e-e6da42b43696'
      insist { subject['source_type'] } == 'RTR'
      insist { subject['hostname'] } == 'product-page-qa-online.dev.cf.i.domain.com'
      insist { subject['date'] } == '[16/12/2014:14:31:55 +0000]'
      insist { subject['verb'] } == 'GET'
      insist { subject['path'] } == '/internal/status'
      insist { subject['http_spec'] } == 'HTTP/1.1'
      insist { subject['http_code'] } == '200'
      insist { subject['size'] } == '6'
      insist { subject['referer'] } == '-'
      insist { subject['user_agent'] } == '-'
      insist { subject['x_forwarded_for'] } == '192.87.127.242, 192.87.127.242, 10.9.0.188, 10.230.16.68'
      insist { subject['vcap_request_id'] } == '5e565011-6935-4afa-5c9f-7597898eec55'
      insist { subject['response_time'] } == '0.005822253'
      insist { subject['message_type'] } == 'OUT'
      insist { subject['source_type'] } == 'RTR'
      insist { subject['source_instance'] } == '0'
    end
  end

  describe "Parse doppler messages with user agent" do
    sample('message' => '{"cf_app_id":"918676fd-0628-422a-a6b2-3aa467e2f0ba","level":"info","message_type":"OUT","msg":"qa-metrics.dev.cf.domain.com - [16/12/2014:14:31:57 +0000] \"GET /radiator HTTP/1.1\" 200 6861 \"-\" \"Mozilla/5.0 (X11; Linux armv6l; rv:24.0) Gecko/20140727 Firefox/24.0 Iceweasel/24.7.0\" 10.230.16.68:47305 x_forwarded_for:\"172.18.11.137\" vcap_request_id:d7fd8c25-4c4b-47f6-6679-abcd67fba02e response_time:0.031576628 app_id:918676fd-0628-422a-a6b2-3aa467e2f0ba\n","source_instance":"0","source_type":"RTR","time":"2014-12-16T15:31:57+01:00"}',
           'program' => 'doppler') do
      insist { subject['user_agent'] } == 'Mozilla/5.0 (X11; Linux armv6l; rv:24.0) Gecko/20140727 Firefox/24.0 Iceweasel/24.7.0'
    end
  end

  describe "We need to support params in the uri path" do
    sample('message' => '{"cf_app_id":"918676fd-0628-422a-a6b2-3aa467e2f0ba","level":"info","message_type":"OUT","msg":"qa-metrics.dev.cf.domain.com - [16/12/2014:14:31:57 +0000] \"GET /api/journal/system/all?callback=functionTestsystem&_=1417096021295 HTTP/1.1\" 200 6861 \"-\" \"Mozilla/5.0 (X11; Linux armv6l; rv:24.0) Gecko/20140727 Firefox/24.0 Iceweasel/24.7.0\" 10.230.16.68:47305 x_forwarded_for:\"172.18.11.137\" vcap_request_id:d7fd8c25-4c4b-47f6-6679-abcd67fba02e response_time:0.031576628 app_id:918676fd-0628-422a-a6b2-3aa467e2f0ba\n","source_instance":"0","source_type":"RTR","time":"2014-12-16T15:31:57+01:00"}',
           'program' => 'doppler') do
      insist { subject['path'] } == '/api/journal/system/all?callback=functionTestsystem&_=1417096021295'
    end
  end

  describe "We clean up unwanted fields" do
    sample('message' => '{"cf_app_id":"f5409d35-755b-4af8-865e-e6da42b43696","level":"info","message_type":"OUT","msg":"product-page-qa-online.dev.cf.i.domain.com - [16/12/2014:14:31:55 +0000] \"GET /internal/status HTTP/1.1\" 200 6 \"-\" \"-\" 10.230.16.68:36015 x_forwarded_for:\"192.87.127.242, 192.87.127.242, 10.9.0.188, 10.230.16.68\" vcap_request_id:5e565011-6935-4afa-5c9f-7597898eec55 response_time:0.005822253 app_id:f5409d35-755b-4af8-865e-e6da42b43696\n","source_instance":"0","source_type":"RTR","time":"2014-12-16T15:31:55+01:00"}',
           'program' => 'doppler') do
      insist { subject['delete_me'] } == nil
      insist { subject['message'] } == nil
      insist { subject['msg'] } == nil
    end
  end
end
