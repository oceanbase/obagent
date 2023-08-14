/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package command

import "errors"

var TimeoutErr = errors.New("command: wait timeout")
var AlreadyFinishedErr = errors.New("command: command already finished")
var ExecutionNotFoundErr = errors.New("command: execution not found")
var ExecutionAlreadyExistsErr = errors.New("command: execution already exists")
var BadTaskFunc = errors.New("command: bad task function")
