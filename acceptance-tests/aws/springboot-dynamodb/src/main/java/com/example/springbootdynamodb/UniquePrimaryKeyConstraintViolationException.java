package com.example.springbootdynamodb;

public class UniquePrimaryKeyConstraintViolationException extends RuntimeException {
    public UniquePrimaryKeyConstraintViolationException(String message) {
        super(message);
    }
}