-- Data transferred during the download and upload phases of each test.
ALTER TABLE speedtest_results ADD COLUMN download_bytes INTEGER;
ALTER TABLE speedtest_results ADD COLUMN upload_bytes INTEGER;
