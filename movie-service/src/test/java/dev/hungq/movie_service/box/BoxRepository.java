package dev.hungq.movie_service.box;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import jakarta.transaction.Transactional;

@Repository
public interface BoxRepository extends JpaRepository<Box, Integer>  
{
    @Transactional
    @Modifying
    @Query("DELETE FROM Box b WHERE b.ownerId = :ownerId")
    void deleteByOwnerId(Integer ownerId);
    
    @Transactional
    @Query(value = "SELECT b.* FROM box_user bx JOIN box b ON b.id = bx.box_id WHERE bx.user_id = :userId", nativeQuery = true)
    Optional<Box> findByUserId(Integer userId);
    
    Optional<Box> findByOwnerId(Integer ownerId);
}
