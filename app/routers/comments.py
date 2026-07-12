from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app import models, schemas, dependencies

router = APIRouter(prefix="/comments", tags=["comments"])

@router.post("/", response_model=schemas.CommentResponse)
def create_comment(comment: schemas.CommentCreate, db: Session = Depends(dependencies.get_db),
                   current_user: models.User = Depends(dependencies.get_current_user)):
    task = db.query(models.Task).filter(models.Task.id == comment.task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    project = db.query(models.Project).filter(models.Project.id == task.project_id).first()
    if current_user.role != models.UserRole.ADMIN and project.owner_id != current_user.id:
        raise HTTPException(status_code=403, detail="Not enough permissions")
    new_comment = models.Comment(text=comment.text, task_id=comment.task_id, author_id=current_user.id)
    db.add(new_comment)
    db.commit()
    db.refresh(new_comment)
    return new_comment

@router.get("/task/{task_id}", response_model=list[schemas.CommentResponse])
def list_comments_for_task(task_id: int, skip: int = 0, limit: int = 50,
                           db: Session = Depends(dependencies.get_db),
                           current_user: models.User = Depends(dependencies.get_current_user)):
    task = db.query(models.Task).filter(models.Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    project = db.query(models.Project).filter(models.Project.id == task.project_id).first()
    if current_user.role != models.UserRole.ADMIN and project.owner_id != current_user.id:
        raise HTTPException(status_code=403, detail="Not enough permissions")
    comments = db.query(models.Comment).filter(models.Comment.task_id == task_id).offset(skip).limit(limit).all()
    return comments