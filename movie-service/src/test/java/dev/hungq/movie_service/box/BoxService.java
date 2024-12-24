package dev.hungq.movie_service.box;

import java.util.List;
import java.util.Optional;
import org.springframework.stereotype.Service;


@Service
public class BoxService {

	private final BoxRepository boxRepo;
	
	public BoxService(BoxRepository boxRepo)
	{
		this.boxRepo = boxRepo;
	}

    public List<Box> findAll() {
        return boxRepo.findAll();
    }

    public Optional<Box> find(Integer id) {
        return boxRepo.findById(id);
    }

    public Box create(Box box) {
        Box b = boxRepo.save(box);
        return b;
    }
    
    public Box save(Box box) {
        return boxRepo.save(box);
    }

    public void delete(Integer id) {
    	boxRepo.deleteById(id);
    }

	public void deleteByOwnerId(Integer ownerId) {
		boxRepo.deleteByOwnerId(ownerId);
	}
	
	public boolean removeUserFromBox(Integer boxId, Integer userId) {
        var obox = boxRepo.findById(boxId);
        
        if (obox.isEmpty())
        	return false;
        
        var box = obox.get();
        if (box.getUserIds().contains(userId)) {
            box.getUserIds().remove(userId);
            boxRepo.save(box);
            return true;
        }
        return false;
	}
	
	public boolean addUserToBox(Integer boxId, Integer userId) {
        var obox = boxRepo.findById(boxId);
        
        if (obox.isEmpty())
        	return false;
        
        var box = obox.get();
        box.getUserIds().add(userId);
        boxRepo.save(box);
        return true;
	}
	
	public boolean containsUser(Integer boxId, Integer userId) {
		var obox = boxRepo.findByUserId(userId);
		return obox.isPresent() && obox.get().getId() == boxId;
	}
	
	public Optional<Box> findByOwnerId(Integer ownerId) {
		return boxRepo.findByOwnerId(ownerId);
	}
	
	
	public Optional<Box> findByUserId(Integer userId) {
		return boxRepo.findByUserId(userId);
	}
}