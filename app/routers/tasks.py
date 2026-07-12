from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app import models, schemas, dependencies

router = APIRouter(prefix="/tasks", tags=["tasks"])

@router.post("/", response_model=schemas.TaskResponse)
def create_task(task: schemas.TaskCreate, db: Session = Depends(dependencies.get_db),
                current_user: models.User = Depends(dependencies.get_current_user)):
    project = db.query(models.Project).filter(models.Project.id == task.project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    if current_user.role != models.UserRole.ADMIN and project.owner_id != current_user.id:
        raise HTTPException(status_code=403, detail="Not enough permissions")
    new_task = models.Task(**task.model_dump(), creator_id=current_user.id)
    db.add(new_task)
    db.commit()
    db.refresh(new_task)
    return new_task

@router.get("/", response_model=list[schemas.TaskResponse])
def list_tasks(
    project_id: int = Query(None),
    status: models.TaskStatus = Query(None),
    priority: models.TaskPriority = Query(None),
    assignee_id: int = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(dependencies.get_db),
    current_user: models.User = Depends(dependencies.get_current_user)
):
    query = db.query(models.Task)
    if project_id:
        query = query.filter(models.Task.project_id == project_id)
    if status:
        query = query.filter(models.Task.status == status)
    if priority:
        query = query.filter(models.Task.priority == priority)
    if assignee_id:
        query = query.filter(models.Task.assignee_id == assignee_id)
    if current_user.role != models.UserRole.ADMIN:
        user_projects = [p.id for p in db.query(models.Project).filter(models.Project.owner_id == current_user.id).all()]
        query = query.filter(models.Task.project_id.in_(user_projects))
    tasks = query.offset(skip).limit(limit).all()
    return tasks

@router.get("/{task_id}", response_model=schemas.TaskResponse)
def get_task(task_id: int, db: Session = Depends(dependencies.get_db),
             current_user: models.User = Depends(dependencies.get_current_user)):
    task = db.query(models.Task).filter(models.Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    project = db.query(models.Project).filter(models.Project.id == task.project_id).first()
    if current_user.role != models.UserRole.ADMIN and project.owner_id != current_user.id:
        raise HTTPException(status_code=403, detail="Not enough permissions")
    return task

@router.put("/{task_id}", response_model=schemas.TaskResponse)
def update_task(task_id: int, task_update: schemas.TaskBase,
                db: Session = Depends(dependencies.get_db),
                current_user: models.User = Depends(dependencies.get_current_user)):
    task = db.query(models.Task).filter(models.Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    project = db.query(models.Project).filter(models.Project.id == task.project_id).first()
    if current_user.role != models.UserRole.ADMIN and project.owner_id != current_user.id:
        raise HTTPException(status_code=403, detail="Not enough permissions")
    for key, value in task_update.model_dump().items():
        setattr(task, key, value)
    db.commit()
    db.refresh(task)
    return task

@router.delete("/{task_id}", status_code=204)
def delete_task(task_id: int, db: Session = Depends(dependencies.get_db),
                current_user: models.User = Depends(dependencies.get_current_user)):
    task = db.query(models.Task).filter(models.Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    project = db.query(models.Project).filter(models.Project.id == task.project_id).first()
    if current_user.role != models.UserRole.ADMIN and project.owner_id != current_user.id:
        raise HTTPException(status_code=403, detail="Not enough permissions")
    db.delete(task)
    db.commit()