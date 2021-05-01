<!--
s3sync-service - Realtime S3 synchronisation tool
Copyright (c) 2020  Yevgeniy Valeyev

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
 -->

This page contains known issues and limitations.

| Description |
|-------------|
| Due to some specifics of the files processing and filepath generation `local_path` should be absolute path to the sync (_site_) directory. So far there's no plans for fixing this. |
| Symlinks within the sync (_site_) directory are ignored since it's hard to properly process them with the current implementation. |
