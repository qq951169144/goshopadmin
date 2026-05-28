# Git 分支合并到 Master 规范流程

## 1. 目的

为避免合并分支到 `master` 时出现工作区状态混乱、文件缺失等问题，制定本规范流程。

## 2. 适用范围

适用于所有从功能分支（如 `feature/*`、`bugfix/*`）合并到 `master` 分支的操作。

## 3. 合并前准备

### 3.1 切换到目标分支并更新

```bash
# 切换到待合并的功能分支
git checkout feature/mq-integration

# 拉取最新代码
git pull origin feature/mq-integration
```

### 3.2 检查工作区状态

**必须确保工作区干净，无未提交的修改**：

```bash
git status
```

**预期输出**：
```
On branch feature/mq-integration
Your branch is up to date with 'origin/feature/mq-integration'.

nothing to commit, working tree clean
```

**如果工作区不干净**：
- 提交修改：`git add . && git commit -m "..."`
- 或暂存修改：`git stash`（合并完成后可通过 `git stash pop` 恢复）

### 3.3 同步工作区（关键步骤）

**强制同步工作区与索引，避免状态混乱**：

```bash
git checkout HEAD -- .
```

**验证同步结果**：
```bash
git status
```

## 4. 执行合并

### 4.1 切换到 master 分支

```bash
git checkout master
```

### 4.2 更新 master 分支

```bash
git pull origin master
```

### 4.3 执行合并

**推荐使用非 Fast-forward 合并**，保留手动解决冲突的机会：

```bash
# 执行合并，遇到冲突时暂停（不自动解决）
git merge --no-ff feature/mq-integration --no-commit
```

**参数说明**：
- `--no-ff`：禁用 Fast-forward，创建合并提交，保留分支历史
- `--no-commit`：合并后不自动提交，允许手动检查和修改

### 4.4 检查冲突并手动解决

合并后检查是否有冲突：

```bash
git status
```

**如果存在冲突**，输出会显示类似：
```
Unmerged paths:
  (use "git add <file>..." to mark resolution)
        both modified:   backend/config/db.go
        both modified:   shop-backend/services/order_service.go
```

**手动解决冲突步骤**：
1. 在 IDE 中打开冲突文件（冲突标记为 `<<<<<<<`, `=======`, `>>>>>>>`）
2. 手动编辑文件，保留需要的代码
3. 保存文件
4. 标记冲突已解决：`git add <冲突文件>`

**查看冲突文件列表**：
```bash
# 列出所有冲突文件
git diff --name-only --diff-filter=U
```

### 4.5 完成合并提交

所有冲突解决后，完成合并提交：

```bash
# 查看工作区状态，确认所有冲突已解决
git status

# 提交合并结果
git commit -m "Merge feature/mq-integration into master"
```

**如果出现文件删除状态（如 `deleted: xxx`）**：

```bash
# 同步工作区
git checkout HEAD -- .

# 再次检查
git status
```

## 5. 合并后验证

### 5.1 验证工作区状态

```bash
git status
```

**预期输出**：
```
On branch master
Your branch is ahead of 'origin/master' by 1 commit.
  (use "git push" to publish your local commits)

nothing to commit, working tree clean
```

### 5.2 验证关键文件存在

```bash
# 验证核心配置文件存在
Test-Path backend/config/db.go
Test-Path shop-backend/config/db.go
```

### 5.3 推送到远程仓库

```bash
git push origin master
```

## 6. 常见问题处理

### 6.1 问题：合并后工作区显示大量 deleted 文件

**原因**：工作区与索引不同步

**解决**：
```bash
# 强制同步工作区
git checkout HEAD -- .

# 验证
git status
```

### 6.2 问题：合并冲突

**处理流程**：
1. 查看冲突文件：`git status` 或 `git diff --name-only --diff-filter=U`
2. 在 IDE 中打开冲突文件（查找 `<<<<<<<`, `=======`, `>>>>>>>` 标记）
3. 手动编辑文件，保留需要的代码
4. 保存文件后标记冲突已解决：`git add <冲突文件>`
5. 完成合并：`git commit -m "Merge feature/xxx into master"`

**取消合并（如果需要）**：
```bash
git merge --abort
```

### 6.3 问题：需要撤销合并

```bash
# 查看合并提交
git log --oneline -5

# 回退到合并前的提交（假设合并提交为 f7cf372）
git reset --hard 2226fcc

# 强制推送到远程
git push -f origin master
```

## 7. 最佳实践

### 7.1 分支命名规范

| 分支类型 | 命名示例 | 说明 |
|---------|---------|------|
| 功能分支 | `feature/mq-integration` | 新功能开发 |
| Bug 修复 | `bugfix/order-delete` | 修复 bug |
| 优化分支 | `optimize/cache` | 代码优化 |

### 7.2 合并频率

- 功能分支开发完成后应及时合并到 `master`
- 避免长期不合并导致分支差异过大

### 7.3 代码审查

合并到 `master` 前建议进行代码审查：
- 检查代码质量
- 验证功能完整性
- 确保测试通过

### 7.4 测试验证

合并前应运行相关测试：
```bash
# 运行单元测试
go test ./...

# 运行集成测试（如有）
go test -v ./services/...
```

## 8. 流程总结

```
┌─────────────────────────────────────────────────────────────┐
│                    合并到 master 流程                        │
├─────────────────────────────────────────────────────────────┤
│  1. git checkout feature/xxx                               │
│  2. git pull origin feature/xxx                            │
│  3. git status (确保干净)                                   │
│  4. git checkout HEAD -- . (同步工作区)                     │
│  5. git checkout master                                    │
│  6. git pull origin master                                 │
│  7. git merge --no-ff feature/xxx --no-commit              │
│  8. git diff --name-only --diff-filter=U (列出冲突文件)     │
│  9. 在 IDE 中手动解决冲突 → git add <冲突文件>              │
│ 10. git commit -m "Merge feature/xxx into master"          │
│ 11. git checkout HEAD -- . (如出现 deleted)                │
│ 12. git push origin master                                 │
└─────────────────────────────────────────────────────────────┘
```

## 9. 注意事项

1. **强制同步**：步骤 4 的 `git checkout HEAD -- .` 是关键步骤，必须执行
2. **冲突处理**：使用 `--no-commit` 参数暂停合并，在 IDE 中手动解决冲突
3. **冲突文件列表**：使用 `git diff --name-only --diff-filter=U` 列出所有冲突文件
4. **冲突标记**：IDE 中查找 `<<<<<<<`, `=======`, `>>>>>>>` 标记定位冲突位置
5. **工作区状态**：合并前后都要检查 `git status`，确保工作区干净
6. **远程推送**：合并完成后及时推送到远程仓库
7. **取消合并**：如果需要取消合并，使用 `git merge --abort`

---

*版本: 1.0*  
*最后更新: 2026-05-28*  
*适用项目: GoShopAdmin*
